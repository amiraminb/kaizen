package cmd

import (
	"github.com/amiraminb/kaizen/internal/repository"
	"github.com/amiraminb/kaizen/internal/service"
	"github.com/spf13/cobra"
)

var (
	fileRepo                       = repository.NewFileRepository()
	repo     repository.Repository = fileRepo
	svc                            = service.New(repo)
)

var version = "dev"

// main prints the error itself, so cobra's duplicate error line and its usage dump
// are suppressed to keep a failure message from being buried in a help wall.
var rootCmd = &cobra.Command{
	Use:           "kaizen",
	Short:         "Local-first habit tracker",
	Version:       version,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() error {
	return rootCmd.Execute()
}
