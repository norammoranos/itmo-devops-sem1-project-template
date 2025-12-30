#!/bin/bash

# Скрипт запуска приложения
# Создаёт VM через Yandex Cloud CLI и разворачивает приложение из Docker Hub

set -e

echo "=== Запуск приложения ===" >&2

# Конфигурация из переменных окружения (они в secrets GitHub Actions)
YC_TOKEN="$YC_TOKEN"
YC_CLOUD_ID="$YC_CLOUD_ID"
YC_FOLDER_ID="$YC_FOLDER_ID"
YC_ZONE="$YC_ZONE"
YC_SUBNET_ID="$YC_SUBNET_ID"
SSH_PUBLIC_KEY="$SSH_PUBLIC_KEY"

DB_HOST="$APP_DB_HOST"
DB_PORT="$APP_DB_PORT"
DB_NAME="$APP_DB_NAME"
DB_USER="$APP_DB_USER"
DB_PASSWORD="$APP_DB_PASSWORD"
APP_PORT="$APP_PORT"

DOCKERHUB_USERNAME="$DOCKERHUB_USERNAME"
VM_NAME="project-sem-1-vm"
DOCKER_IMAGE="${DOCKERHUB_USERNAME}/project-sem-1:latest"

# Настройка Yandex Cloud CLI
echo "Настройка Yandex Cloud CLI..." >&2
yc config set token "$YC_TOKEN" >&2
yc config set cloud-id "$YC_CLOUD_ID" >&2
yc config set folder-id "$YC_FOLDER_ID" >&2

# Проверка существования VM
echo "Проверка существования VM..." >&2
EXISTING_VM=$(yc compute instance list --format json | jq -r ".[] | select(.name==\"$VM_NAME\") | .id")

if [ -n "$EXISTING_VM" ]; then
    echo "VM уже существует, удаляем..." >&2
    yc compute instance delete --id "$EXISTING_VM" >&2
    sleep 10
fi

# Создание VM
echo "Создание VM через Yandex Cloud..." >&2
VM_INFO=$(yc compute instance create \
    --name "$VM_NAME" \
    --zone "$YC_ZONE" \
    --network-interface "subnet-id=$YC_SUBNET_ID,nat-ip-version=ipv4" \
    --create-boot-disk "image-folder-id=standard-images,image-family=ubuntu-2204-lts,size=20" \
    --memory 2G \
    --cores 2 \
    --ssh-key "$SSH_PUBLIC_KEY" \
    --format json)

VM_IP=$(echo "$VM_INFO" | jq -r '.network_interfaces[0].primary_v4_address.one_to_one_nat.address')

echo "VM создана с IP: $VM_IP" >&2

# Ожидание готовности VM
echo "Ожидание готовности VM..." >&2
sleep 60

for i in {1..30}; do
    if ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 yc-user@"$VM_IP" "exit 0" 2>/dev/null; then
        echo "VM готова!" >&2
        break
    fi
    echo "Ожидание SSH... ($i/30)" >&2
    sleep 10
done

# Копирование docker-compose.yaml
scp -o StrictHostKeyChecking=no docker-compose.yaml yc-user@"$VM_IP":/tmp/ >&2

# Установка Docker и запуск приложения на VM
echo "Установка Docker и запуск приложения..." >&2
ssh -o StrictHostKeyChecking=no yc-user@"$VM_IP" "
    set -e

    echo '=== Установка Docker ===' >&2
    sudo apt-get update -qq
    sudo apt-get install -y -qq ca-certificates curl gnupg
    sudo install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    sudo chmod a+r /etc/apt/keyrings/docker.gpg
    echo \"deb [arch=\$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \$(. /etc/os-release && echo \$VERSION_CODENAME) stable\" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    sudo apt-get update -qq
    sudo apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-compose-plugin
    sudo systemctl start docker
    sudo systemctl enable docker
    sudo usermod -aG docker yc-user

    echo '=== Загрузка Docker-образа из Docker Hub ===' >&2
    sudo docker pull ${DOCKER_IMAGE}

    echo '=== Настройка переменных окружения ===' >&2
    cat > /tmp/.env <<EOF
DOCKERHUB_USERNAME=${DOCKERHUB_USERNAME}
APP_DB_HOST=database
APP_DB_PORT=${DB_PORT}
APP_DB_NAME=${DB_NAME}
APP_DB_USER=${DB_USER}
APP_DB_PASSWORD=${DB_PASSWORD}
APP_PORT=${APP_PORT}
POSTGRES_DB=${DB_NAME}
POSTGRES_USER=${DB_USER}
POSTGRES_PASSWORD=${DB_PASSWORD}
EOF

    echo '=== Запуск Docker Compose ===' >&2
    cd /tmp
    sudo docker compose --env-file .env up -d --no-build

    echo '=== Ожидание запуска сервисов ===' >&2
    sleep 10

    echo '=== Проверка статуса ===' >&2
    sudo docker compose ps >&2

    if sudo docker compose ps | grep -q 'Up\|running'; then
        echo 'Приложение успешно запущено!' >&2
    else
        echo 'Ошибка запуска приложения' >&2
        sudo docker compose logs >&2
        exit 1
    fi
" >&2

echo "=== Приложение запущено ===" >&2
echo "$VM_IP"
