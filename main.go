package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"gopkg.in/yaml.v3"
)

type DomainConfig struct {
	BaseRedirectURL    string `yaml:"base_redirect_url"`
	ExpectBotParam     string `yaml:"expect_bot_param"`
	ExpectBotValue     string `yaml:"expect_bot_value"`
	BotCookieName      string `yaml:"bot_cookie_name"`
	BotCookieValue     string `yaml:"bot_cookie_value"`
	PageTemplate       string `yaml:"page_template"`
	MinRedirectSeconds int    `yaml:"min_redirect_seconds"`
	MaxRedirectSeconds int    `yaml:"max_redirect_seconds"`
}

type Config struct {
	Port    int                     `yaml:"port"`
	Domains map[string]DomainConfig `yaml:"domains"`
}

type ConfigManager struct {
	config     *Config
	configPath string
	mutex      sync.RWMutex
}

func NewConfigManager(configPath string) (*ConfigManager, error) {
	cm := &ConfigManager{
		configPath: configPath,
	}

	if err := cm.loadConfig(); err != nil {
		return nil, err
	}

	if err := cm.watchConfig(); err != nil {
		return nil, err
	}

	return cm, nil
}

func (cm *ConfigManager) GetConfig() Config {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return *cm.config
}

func (cm *ConfigManager) GetDomainConfig(domain string) (*DomainConfig, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if config, exists := cm.config.Domains[domain]; exists {
		return &config, nil
	}

	for d, config := range cm.config.Domains {
		if strings.HasSuffix(domain, "."+d) {
			return &config, nil
		}
	}

	return nil, fmt.Errorf("no configuration found for domain: %s", domain)
}

func (cm *ConfigManager) loadConfig() error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	cm.mutex.Lock()
	cm.config = &config
	cm.mutex.Unlock()

	log.Println("Configuration loaded successfully")
	return nil
}

func (cm *ConfigManager) watchConfig() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating file watcher: %w", err)
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					log.Println("Config file changed, reloading...")
					if err := cm.loadConfig(); err != nil {
						log.Printf("Error reloading config: %v\n", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v\n", err)
			}
		}
	}()

	if err := watcher.Add(cm.configPath); err != nil {
		return fmt.Errorf("error watching config file: %w", err)
	}

	log.Printf("Started watching config file: %s\n", cm.configPath)
	return nil
}

func getDomainFromHost(host string) string {
	host = strings.Split(host, ":")[0]

	if host == "localhost" || net.ParseIP(host) != nil {
		return host
	}

	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return host
}

func main() {
	configPath := "config.yaml"

	configManager, err := NewConfigManager(configPath)
	if err != nil {
		log.Fatalf("Error initializing config manager: %v\n", err)
	}

	engine := html.New("./views", ".gohtml")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(func(c *fiber.Ctx) error {
		domain := getDomainFromHost(c.Hostname())

		domainConfig, err := configManager.GetDomainConfig(domain)
		if err != nil {
			log.Printf("Error getting domain config for %s: %v\n", domain, err)
			return c.SendStatus(fiber.StatusNotFound)
		}

		c.Locals("domainConfig", domainConfig)
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		domainConfig := c.Locals("domainConfig").(*DomainConfig)

		botValue := c.Query(domainConfig.ExpectBotParam)
		cookie := c.Cookies(domainConfig.BotCookieName)

		if botValue != domainConfig.ExpectBotValue && cookie != domainConfig.BotCookieValue {
			return c.Redirect(domainConfig.BaseRedirectURL)
		}

		c.Cookie(&fiber.Cookie{
			Name:    domainConfig.BotCookieName,
			Value:   domainConfig.BotCookieValue,
			Expires: time.Now().Add(365 * 24 * time.Hour),
		})

		return c.Render("index", fiber.Map{
			"PageTemplate":       domainConfig.PageTemplate,
			"MinRedirectSeconds": domainConfig.MinRedirectSeconds,
			"MaxRedirectSeconds": domainConfig.MaxRedirectSeconds,
		})
	})

	app.Get("/index:id.html", func(c *fiber.Ctx) error {
		domainConfig := c.Locals("domainConfig").(*DomainConfig)

		id := c.Params("id")

		_, err := strconv.Atoi(id)
		if err != nil {
			return c.Redirect("/")
		}

		botValue := c.Query(domainConfig.ExpectBotParam)
		cookie := c.Cookies(domainConfig.BotCookieName)

		if botValue != domainConfig.ExpectBotValue && cookie != domainConfig.BotCookieValue {
			return c.Redirect(domainConfig.BaseRedirectURL)
		}

		c.Cookie(&fiber.Cookie{
			Name:    domainConfig.BotCookieName,
			Value:   domainConfig.BotCookieValue,
			Expires: time.Now().Add(365 * 24 * time.Hour),
		})

		return c.Render("index", fiber.Map{
			"PageTemplate":       domainConfig.PageTemplate,
			"MinRedirectSeconds": domainConfig.MinRedirectSeconds,
			"MaxRedirectSeconds": domainConfig.MaxRedirectSeconds,
		})
	})

	port := configManager.GetConfig().Port
	if err := app.Listen(":" + strconv.Itoa(port)); err != nil {
		log.Fatalln("Failed to listen on server: ", err)
	}
}
