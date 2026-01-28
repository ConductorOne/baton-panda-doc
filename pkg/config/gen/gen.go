package main

import (
	cfg "github.com/conductorone/baton-panda-doc/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("panda-doc", cfg.Config)
}
