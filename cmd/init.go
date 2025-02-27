package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

var initTemplate string

var initCmd = &cobra.Command{
	Use:     "init [DIR]",
	Short:   "Init a leetcode workspace",
	Example: "leetgo init -t us -l cpp",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}
		if initTemplate != "" && initTemplate != "us" && initTemplate != "cn" {
			return fmt.Errorf("invalid template %s, only us or cn is supported", initTemplate)
		}
		err := utils.CreateIfNotExists(dir, true)
		if err != nil {
			return err
		}
		err = createConfigDir()
		if err != nil {
			return err
		}
		err = createConfigFiles(dir)
		if err != nil {
			return err
		}
		err = createQuestionCache()
		return err
	},
}

func init() {
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "us", "template to use, cn or us")
}

func createConfigDir() error {
	dir := config.Get().ConfigDir()
	if utils.IsExist(dir) {
		return nil
	}
	err := utils.MakeDir(dir)
	if err != nil {
		return err
	}
	log.Info("config dir created", "dir", dir)
	return nil
}

func createConfigFiles(dir string) error {
	cfg := config.Get()
	site := cfg.LeetCode.Site
	language := cfg.Language
	if initTemplate == "us" {
		site = config.LeetCodeUS
		language = config.EN
	} else if initTemplate == "cn" {
		site = config.LeetCodeCN
		language = config.ZH
	}

	globalFile := cfg.GlobalConfigFile()
	if !utils.IsExist(globalFile) {
		f, err := os.Create(globalFile)
		if err != nil {
			return err
		}

		cfg.LeetCode.Site = site
		cfg.Language = language

		err = cfg.Write(f, true)
		if err != nil {
			return err
		}
		log.Info("global config file created", "file", globalFile)
	}

	projectFile := filepath.Join(dir, config.ProjectConfigFilename)
	f, err := os.Create(projectFile)
	if err != nil {
		return err
	}

	tmpl := `# leetgo project level config, global config is at %s
# for more details, please refer to https://github.com/j178/leetgo
language: %s
code:
  lang: %s
leetcode:
  site: %s
#  credentials:
#    from: browser
#editor:
#  use: none
`
	_, _ = f.WriteString(
		fmt.Sprintf(
			tmpl,
			globalFile,
			language,
			cfg.Code.Lang,
			site,
		),
	)
	log.Info("project config file created", "file", projectFile)

	return nil
}

func createQuestionCache() error {
	c := leetcode.NewClient()
	cache := leetcode.GetCache(c)
	if !cache.Outdated() {
		return nil
	}
	err := cache.Update()
	if err != nil {
		return err
	}
	return nil
}
