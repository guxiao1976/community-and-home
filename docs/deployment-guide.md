# Docker 部署配置指南

本文档说明如何配置 `.github/workflows/docker-deploy.yml` 进行实际部署。

## 📋 当前状态

✅ **已完成**：
- Docker 镜像构建 workflow
- 多服务并行构建（matrix strategy）
- 自动推送到 GitHub Container Registry (ghcr.io)
- 镜像标签策略（branch、tag、sha）
- 构建缓存优化

⚠️ **待配置**：
- 部署到 Kubernetes/云平台
- 健康检查脚本
- 回滚机制
- 环境变量和 Secrets

---

## 🚀 快速开始

### 1. 配置镜像仓库

**GitHub Container Registry（推荐，已配置）**：
- 自动使用 `ghcr.io`
- 无需额外配置
- 镜像地址：`ghcr.io/<your-org>/<service-name>:<tag>`

**其他仓库（如 Docker Hub、阿里云 ACR）**：
```yaml
# 修改 docker-deploy.yml 的 env 部分
env:
  REGISTRY: docker.io  # 或 registry.cn-hangzhou.aliyuncs.com
  IMAGE_PREFIX: your-username
```

然后添加认证 Secret：
```bash
# GitHub Settings → Secrets and variables → Actions
DOCKER_USERNAME=your-username
DOCKER_PASSWORD=your-password
```

### 2. 启用 GitHub Container Registry

如果使用 ghcr.io（默认配置）：

1. **首次推送后设置镜像为 Public**（可选）：
   - 访问 `https://github.com/orgs/<your-org>/packages`
   - 找到服务镜像 → Package settings → Change visibility

2. **验证权限**：
   - Workflow 自动使用 `GITHUB_TOKEN`
   - 无需手动配置

---

## 🎯 部署配置

### 选项 1：Kubernetes 部署

#### 前提条件
- Kubernetes 集群（EKS、GKE、自建等）
- `kubectl` 配置文件

#### 配置步骤

**1. 添加 Kubernetes 配置到 GitHub Secrets**：
```bash
# 方法 1：直接上传 kubeconfig
cat ~/.kube/config | base64

# 在 GitHub Secrets 中创建 KUBE_CONFIG_DATA
# 内容为上面的 base64 字符串
```

**2. 取消注释 `deploy-staging` job 并修改**：
```yaml
deploy-staging:
  name: Deploy to Staging
  runs-on: ubuntu-latest
  needs: build
  if: github.ref == 'refs/heads/develop'
  environment:
    name: staging
    url: https://staging.your-domain.com

  steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Setup kubectl
      uses: azure/setup-kubectl@v3

    - name: Configure kubectl
      run: |
        mkdir -p ~/.kube
        echo "${{ secrets.KUBE_CONFIG_DATA }}" | base64 -d > ~/.kube/config
        kubectl config use-context staging-cluster

    - name: Deploy services
      run: |
        # 更新镜像
        for service in user-service auth-service permission-service; do
          kubectl set image deployment/$service \
            $service=${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}/$service:${{ github.sha }} \
            -n your-namespace
          
          # 等待 rollout 完成
          kubectl rollout status deployment/$service -n your-namespace --timeout=5m
        done

    - name: Health check
      run: |
        # 检查所有 Pod 是否健康
        kubectl get pods -n your-namespace
        
        # 调用健康检查接口
        for service in user-service auth-service; do
          kubectl run curl-test --image=curlimages/curl --rm -it --restart=Never -- \
            curl -f http://$service.your-namespace.svc.cluster.local:8080/health
        done
```

**3. 配置回滚**：
```yaml
    - name: Rollback on failure
      if: failure()
      run: |
        echo "⚠️ Deployment failed, rolling back..."
        for service in user-service auth-service permission-service; do
          kubectl rollout undo deployment/$service -n your-namespace
          kubectl rollout status deployment/$service -n your-namespace
        done
```

### 选项 2：Docker Compose 部署

适用于单机或小规模部署。

**1. 配置 SSH 访问**：
```bash
# 在 GitHub Secrets 中添加
SSH_PRIVATE_KEY=<your-private-key>
SSH_HOST=your-server.com
SSH_USER=deploy
```

**2. 部署脚本**：
```yaml
deploy-staging:
  name: Deploy to Staging
  runs-on: ubuntu-latest
  needs: build
  if: github.ref == 'refs/heads/develop'

  steps:
    - name: Deploy via SSH
      uses: appleboy/ssh-action@v1.0.0
      with:
        host: ${{ secrets.SSH_HOST }}
        username: ${{ secrets.SSH_USER }}
        key: ${{ secrets.SSH_PRIVATE_KEY }}
        script: |
          cd /opt/community-home
          
          # 拉取最新镜像
          docker compose pull
          
          # 滚动更新
          docker compose up -d --no-deps --build user-service
          docker compose up -d --no-deps --build auth-service
          
          # 健康检查
          sleep 10
          curl -f http://localhost:8001/health || exit 1
          curl -f http://localhost:8002/health || exit 1
```

### 选项 3：云平台部署

#### AWS ECS
使用 `aws-actions/amazon-ecs-deploy-task-definition@v1`

#### Google Cloud Run
使用 `google-github-actions/deploy-cloudrun@v2`

#### 阿里云 ACK
参考 Kubernetes 部署方式

---

## 🔒 环境变量和 Secrets

### 必需的 Secrets（根据选项）

**Kubernetes 部署**：
- `KUBE_CONFIG_DATA` - Kubernetes 配置（base64）

**SSH 部署**：
- `SSH_PRIVATE_KEY` - SSH 私钥
- `SSH_HOST` - 服务器地址
- `SSH_USER` - SSH 用户名

**自定义镜像仓库**：
- `DOCKER_USERNAME` - 镜像仓库用户名
- `DOCKER_PASSWORD` - 镜像仓库密码

### 添加 Secret 步骤

1. 访问 GitHub 仓库
2. Settings → Secrets and variables → Actions
3. New repository secret
4. 输入名称和值
5. Add secret

---

## ✅ 验证部署

### 本地测试构建

```bash
# 测试单个服务的 Docker 构建
cd services/user-service
docker build -t user-service:test .

# 测试运行
docker run --rm -p 8001:8001 user-service:test
```

### 触发 Workflow

**方式 1：推送到 main/master**
```bash
git push origin main
```

**方式 2：打 tag 触发生产部署**
```bash
git tag v1.0.0
git push origin v1.0.0
```

**方式 3：手动触发**
- GitHub → Actions → Docker Build and Deploy
- Run workflow → 选择环境

---

## 📊 监控部署状态

### GitHub Actions 界面
查看实时日志和部署状态

### Kubernetes 监控
```bash
# 查看 Pod 状态
kubectl get pods -n your-namespace

# 查看最近的 events
kubectl get events -n your-namespace --sort-by='.lastTimestamp'

# 查看服务日志
kubectl logs -f deployment/user-service -n your-namespace
```

---

## 🔄 回滚策略

### 自动回滚
已配置在 workflow 中，健康检查失败时自动触发

### 手动回滚

**Kubernetes**：
```bash
# 回滚到上一个版本
kubectl rollout undo deployment/user-service -n your-namespace

# 回滚到指定版本
kubectl rollout undo deployment/user-service --to-revision=2 -n your-namespace
```

**Docker Compose**：
```bash
# 拉取之前的镜像版本
docker pull ghcr.io/your-org/user-service:previous-sha
docker compose up -d user-service
```

---

## 🎯 下一步

1. ✅ 确定部署目标（Kubernetes/Docker Compose/云平台）
2. ✅ 配置必要的 Secrets
3. ✅ 取消注释 `deploy-staging` 或 `deploy-production` job
4. ✅ 根据实际情况修改部署脚本
5. ✅ 配置健康检查和回滚逻辑
6. ✅ 在测试环境验证
7. ✅ 配置生产环境部署

---

## 📚 参考资料

- [GitHub Actions 文档](https://docs.github.com/actions)
- [GitHub Container Registry 文档](https://docs.github.com/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Kubernetes 部署文档](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Docker Compose 文档](https://docs.docker.com/compose/)

---

**需要帮助？** 请参考项目文档或联系 DevOps 团队。
