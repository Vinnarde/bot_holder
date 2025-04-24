package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"gopkg.in/yaml.v3"
)

type ConfigManager struct {
	config     *Config
	configPath string
	mutex      sync.RWMutex
}

type Config struct {
	Port               int    `yaml:"port"`
	BaseRedirectURL    string `yaml:"base_redirect_url"`
	ExpectBotParam     string `yaml:"expect_bot_param"`
	ExpectBotValue     string `yaml:"expect_bot_value"`
	BotCookieName      string `yaml:"bot_cookie_name"`
	BotCookieValue     string `yaml:"bot_cookie_value"`
	PageTemplate       string `yaml:"page_template"`
	MinRedirectSeconds int    `yaml:"min_redirect_seconds"`
	MaxRedirectSeconds int    `yaml:"max_redirect_seconds"`
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
				// Check if the config file was modified
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

	app.Get("/", func(c *fiber.Ctx) error {
		config := configManager.GetConfig()

		botValue := c.Query(config.ExpectBotParam)
		cookie := c.Cookies(config.BotCookieName)

		if botValue != config.ExpectBotValue && cookie != config.BotCookieValue {
			return c.Redirect(config.BaseRedirectURL)
		}

		c.Cookie(&fiber.Cookie{
			Name:    config.BotCookieName,
			Value:   config.BotCookieValue,
			Expires: time.Now().Add(365 * 24 * time.Hour),
		})

		return c.Render("index", fiber.Map{
			"PageTemplate":       config.PageTemplate,
			"MinRedirectSeconds": config.MinRedirectSeconds,
			"MaxRedirectSeconds": config.MaxRedirectSeconds,
		})
	})

	app.Get("/index:id.html", func(c *fiber.Ctx) error {
		config := configManager.GetConfig()

		id := c.Params("id")
		_, err := strconv.Atoi(id)
		if err != nil {
			return c.Redirect("/")
		}

		botValue := c.Query(config.ExpectBotParam)
		cookie := c.Cookies(config.BotCookieName)

		if botValue != config.ExpectBotValue && cookie != config.BotCookieValue {
			return c.Redirect(config.BaseRedirectURL)
		}

		c.Cookie(&fiber.Cookie{
			Name:    config.BotCookieName,
			Value:   config.BotCookieValue,
			Expires: time.Now().Add(365 * 24 * time.Hour),
		})

		return c.Render("index", fiber.Map{
			"PageTemplate":       config.PageTemplate,
			"MinRedirectSeconds": config.MinRedirectSeconds,
			"MaxRedirectSeconds": config.MaxRedirectSeconds,
		})
	})

	app.Get("/status", func(c *fiber.Ctx) error {
		config := configManager.GetConfig()
		return c.JSON(fiber.Map{
			"status": "running",
			"config": fiber.Map{
				"port":               config.Port,
				"baseRedirectUrl":    config.BaseRedirectURL,
				"minRedirectSeconds": config.MinRedirectSeconds,
				"maxRedirectSeconds": config.MaxRedirectSeconds,
				"pageTemplate":       config.PageTemplate,
			},
		})
	})

	port := configManager.GetConfig().Port
	if err := app.Listen(":" + strconv.Itoa(port)); err != nil {
		log.Fatalln("Failed to listen on server: ", err)
	}
}
