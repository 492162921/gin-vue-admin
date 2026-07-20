package inspection

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
	inspRequest "github.com/flipped-aurora/gin-vue-admin/server/model/inspection/request"
	"github.com/flipped-aurora/gin-vue-admin/server/service/inspection/k8s"
)

type ClusterService struct{}

func (s *ClusterService) Create(ctx context.Context, name string, kubeconfigBytes []byte) (*inspModel.InspCluster, error) {
	path, err := saveKubeconfig(kubeconfigBytes)
	if err != nil {
		return nil, err
	}
	cluster := inspModel.InspCluster{Name: name, KubeconfigPath: path, Status: "unknown"}
	if err := s.probe(ctx, &cluster); err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&cluster).Error; err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	return &cluster, nil
}

func (s *ClusterService) Refresh(ctx context.Context, id uint) (*inspModel.InspCluster, error) {
	cluster, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.probe(ctx, cluster); err != nil {
		if saveErr := global.GVA_DB.WithContext(ctx).Save(cluster).Error; saveErr != nil {
			return nil, saveErr
		}
		return nil, err
	}
	if err := global.GVA_DB.WithContext(ctx).Save(cluster).Error; err != nil {
		return nil, err
	}
	return cluster, nil
}

func (s *ClusterService) List(ctx context.Context, search inspRequest.ClusterSearch) ([]inspModel.InspCluster, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&inspModel.InspCluster{})
	if search.Name != "" {
		db = db.Where("name LIKE ?", "%"+search.Name+"%")
	}
	if search.Status != "" {
		db = db.Where("status = ?", search.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	var clusters []inspModel.InspCluster
	if err := db.Order("id DESC").Limit(limit).Offset(offset).Find(&clusters).Error; err != nil {
		return nil, 0, err
	}
	return clusters, total, nil
}

func (s *ClusterService) Get(ctx context.Context, id uint) (*inspModel.InspCluster, error) {
	var cluster inspModel.InspCluster
	if err := global.GVA_DB.WithContext(ctx).First(&cluster, id).Error; err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (s *ClusterService) Update(ctx context.Context, id uint, input inspRequest.UpdateCluster) (*inspModel.InspCluster, error) {
	cluster, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	cluster.Name = input.Name
	if input.Kubeconfig != "" {
		path, err := saveKubeconfig([]byte(input.Kubeconfig))
		if err != nil {
			return nil, err
		}
		oldPath := cluster.KubeconfigPath
		cluster.KubeconfigPath = path
		if err := s.probe(ctx, cluster); err != nil {
			_ = os.Remove(path)
			return nil, err
		}
		if err := global.GVA_DB.WithContext(ctx).Save(cluster).Error; err != nil {
			_ = os.Remove(path)
			return nil, err
		}
		if oldPath != "" && oldPath != path {
			_ = os.Remove(oldPath)
		}
		return cluster, nil
	}
	if err := global.GVA_DB.WithContext(ctx).Save(cluster).Error; err != nil {
		return nil, err
	}
	return cluster, nil
}

func (s *ClusterService) Delete(ctx context.Context, id uint) error {
	cluster, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := global.GVA_DB.WithContext(ctx).Delete(cluster).Error; err != nil {
		return err
	}
	if cluster.KubeconfigPath != "" {
		_ = os.Remove(cluster.KubeconfigPath)
	}
	return nil
}

func (s *ClusterService) ListNodes(ctx context.Context, id uint) ([]k8s.NodeInfo, error) {
	cluster, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return k8s.NewInspector(cluster.KubeconfigPath, global.GVA_CONFIG.Inspection.ForceStub).ListNodes(ctx)
}

func (s *ClusterService) ListNamespaces(ctx context.Context, id uint) ([]k8s.NamespaceInfo, error) {
	cluster, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return k8s.NewInspector(cluster.KubeconfigPath, global.GVA_CONFIG.Inspection.ForceStub).ListNamespaces(ctx)
}

func (s *ClusterService) probe(ctx context.Context, cluster *inspModel.InspCluster) error {
	inspector := k8s.NewInspector(cluster.KubeconfigPath, global.GVA_CONFIG.Inspection.ForceStub)
	if err := inspector.TestConnection(ctx); err != nil {
		cluster.Status = "unreachable"
		cluster.K8sVersion = ""
		cluster.NodeCount = 0
		return fmt.Errorf("集群连接或 kubeconfig 验证失败: %w", err)
	}
	cluster.Status = "available"
	cluster.K8sVersion, _ = inspector.GetVersion(ctx)
	nodes, err := inspector.ListNodes(ctx)
	if err == nil {
		cluster.NodeCount = len(nodes)
	}
	return nil
}

func saveKubeconfig(contents []byte) (string, error) {
	if len(contents) == 0 {
		return "", fmt.Errorf("kubeconfig 不能为空")
	}
	dir := global.GVA_CONFIG.Inspection.KubeConfigDir
	if dir == "" {
		dir = "uploads/kubeconfigs"
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	path := filepath.Join(dir, hex.EncodeToString(token)+".yaml")
	if err := os.WriteFile(path, contents, 0600); err != nil {
		return "", err
	}
	return path, nil
}
