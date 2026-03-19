#!/bin/bash
# WSL Ubuntu 22.04.5 開発環境セットアップスクリプト
# 実行: bash scripts/setup-wsl.sh
# プロジェクトルートで実行（WSL にコピー後: ~/work/AI/project_management_tool、または /mnt/c/... から直接）

set -e

echo "=== WSL 開発環境セットアップ ==="

# 1. パッケージ更新
echo "[1/6] パッケージ更新..."
sudo apt update && sudo apt upgrade -y

# 2. 基本ツール
echo "[2/6] Git インストール..."
sudo apt install -y git ca-certificates curl gnupg lsb-release

# 3. Docker Engine
echo "[3/6] Docker Engine インストール..."
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker "$USER"
echo "  -> ログアウト/再ログイン後、docker が sudo なしで使えます"

# 4. Go
echo "[4/6] Go インストール..."
GO_VERSION="1.22.4"
GO_ARCH="linux-amd64"
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz" -o /tmp/go.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf /tmp/go.tar.gz
rm -f /tmp/go.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo "export PATH=\$PATH:$(/usr/local/go/bin/go env GOPATH)/bin" >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin
echo "  -> Go ${GO_VERSION} をインストールしました"

# 5. Node.js (nvm)
echo "[5/6] Node.js (nvm) インストール..."
if [ ! -d "$HOME/.nvm" ]; then
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
  export NVM_DIR="$HOME/.nvm"
  [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
fi
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
nvm install 20
nvm use 20
nvm alias default 20
echo "  -> Node.js 20 をインストールしました"

# 6. GitHub CLI & GCP CLI
echo "[6/6] GitHub CLI / GCP CLI インストール..."
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/etc/apt/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install -y gh

# GCP CLI
if ! command -v gcloud &> /dev/null; then
  curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-linux-x86_64.tar.gz
  tar -xf google-cloud-cli-linux-x86_64.tar.gz -C "$HOME"
  "$HOME/google-cloud-sdk/install.sh" --quiet --path-update true
  rm -f google-cloud-cli-linux-x86_64.tar.gz
  echo "  -> GCP CLI をインストールしました"
else
  echo "  -> GCP CLI は既にインストール済み"
fi

echo ""
echo "=== セットアップ完了 ==="
echo ""
echo "次のステップ:"
echo "  1. 新しいターミナルを開く（または wsl --shutdown 後、Ubuntu を再起動）"
echo "  2. Docker を起動: sudo service docker start (または sudo systemctl start docker)"
echo "  3. プロジェクトを clone（未実施の場合）: mkdir -p ~/work/AI && cd ~/work/AI && git clone git@github.com:uraguchihiroki/project_management_tool.git"
echo "  4. Cursor で WSL に接続し、~/work/AI/project_management_tool を開く"
echo "  5. bash scripts/start.sh でアプリを起動"
echo ""
