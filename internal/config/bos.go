package config

import (
	"fmt"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos"
)

func InitBOS(cfg BOSConfig) (*bos.Client, error) {
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		// 开发环境允许空配置
		return nil, nil
	}

	clientConfig := bos.BosClientConfiguration{
		Ak:               cfg.AccessKeyID,
		Sk:               cfg.SecretAccessKey,
		Endpoint:         cfg.Endpoint,
		RedirectDisabled: false,
	}

	client, err := bos.NewClientWithConfig(&clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create BOS client: %w", err)
	}

	client.Config.Retry = bce.NewBackOffRetryPolicy(3, 20000, 300)

	return client, nil
}
