package svc

import (
	"context"
	"log"

	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	"github.com/guxiao1976/community-file/model"
	"github.com/guxiao1976/community-file/rpc/internal/config"
	"github.com/guxiao1976/community-common/v2/pkg/minio"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	gominio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	SysConfig *sysconfig.Client
	FileModel model.FileModel
	MinioCli  *minio.Client       // common 封装
	RawMinio  *gominio.Client     // 原始客户端（用于 PresignedPutObject）
	Bucket    string
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := sqlx.NewMysql(c.DataSource.Source)

	minioCli, err := minio.NewClient(minio.Config{
		Endpoint:        c.Minio.Endpoint,
		AccessKeyID:     c.Minio.AccessKey,
		SecretAccessKey: c.Minio.SecretKey,
		BucketName:      c.Minio.Bucket,
		UseSSL:          c.Minio.UseSSL,
	})
	if err != nil {
		log.Printf("Warning: MinIO client init failed: %v", err)
	}

	rawCli, err := gominio.New(c.Minio.Endpoint, &gominio.Options{
		Creds:  credentials.NewStaticV4(c.Minio.AccessKey, c.Minio.SecretKey, ""),
		Secure: c.Minio.UseSSL,
	})
	if err != nil {
		log.Printf("Warning: Raw MinIO client init failed: %v", err)
	} else {
		// ensure bucket exists
		exists, _ := rawCli.BucketExists(context.Background(), c.Minio.Bucket)
		if !exists {
			rawCli.MakeBucket(context.Background(), c.Minio.Bucket, gominio.MakeBucketOptions{})
		}
	}

	// 初始化系统参数配置客户端
	sysCfg := sysconfig.MustInit(c.SysConfigRedis, "", func(ctx context.Context, key string) (*sysconfig.ConfigValue, error) {
		// gRPC fallback to master-data GetConfig
		conn, err := zrpc.NewClient(c.MasterDataRpc)
		if err != nil {
			return nil, err
		}
		client := masterdatav1.NewMasterdataServiceClient(conn.Conn())
		resp, err := client.GetConfig(ctx, &masterdatav1.GetConfigReq{ConfigKey: key})
		if err != nil {
			return nil, err
		}
		return &sysconfig.ConfigValue{
			Value: resp.ConfigValue,
			Type:  resp.ValueType,
			Desc:  resp.Description,
		}, nil
	})

	return &ServiceContext{
		Config:    c,
		SysConfig: sysCfg,
		FileModel: model.NewFileModel(db),
		MinioCli:  minioCli,
		RawMinio:  rawCli,
		Bucket:    c.Minio.Bucket,
	}
}
