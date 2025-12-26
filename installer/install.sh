#!/bin/bash
# ============================================
# COLD STORAGE - ONE-CLICK DISASTER RECOVERY
# ============================================
# Usage: ./install.sh
# This script sets up everything automatically
#
# Supported distros:
#   - Ubuntu/Debian (apt)
#   - Arch Linux (pacman)
#   - Fedora/RHEL/CentOS (dnf/yum)
#   - openSUSE (zypper)

set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║     COLD STORAGE - DISASTER RECOVERY INSTALLER             ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get script directory (where binary should be)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BINARY="$SCRIPT_DIR/server"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root: sudo ./install.sh${NC}"
    exit 1
fi

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo -e "${RED}Error: server binary not found in $SCRIPT_DIR${NC}"
    echo "Make sure 'server' binary is in the same folder as this script"
    exit 1
fi

# ============================================
# DETECT PACKAGE MANAGER
# ============================================
detect_distro() {
    # Check for macOS first
    if [ "$(uname)" = "Darwin" ]; then
        echo "macos"
        return
    fi

    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO=$ID
    elif [ -f /etc/arch-release ]; then
        DISTRO="arch"
    elif [ -f /etc/debian_version ]; then
        DISTRO="debian"
    elif [ -f /etc/redhat-release ]; then
        DISTRO="rhel"
    else
        DISTRO="unknown"
    fi
    echo $DISTRO
}

DISTRO=$(detect_distro)
echo -e "${GREEN}Detected distro:${NC} $DISTRO"

# ============================================
# PACKAGE MANAGER FUNCTIONS
# ============================================
install_postgresql() {
    case $DISTRO in
        macos)
            echo -e "${GREEN}[1/5]${NC} Checking Homebrew..."
            if ! command -v brew &> /dev/null; then
                echo -e "${RED}Homebrew not found. Install it first:${NC}"
                echo '  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'
                exit 1
            fi
            echo -e "${GREEN}[2/5]${NC} Installing PostgreSQL..."
            if command -v psql &> /dev/null; then
                echo "  PostgreSQL already installed"
                brew services start postgresql@16 2>/dev/null || brew services start postgresql 2>/dev/null || true
            else
                brew install postgresql@16
                brew services start postgresql@16
                echo "  PostgreSQL installed and started"
            fi
            ;;
        ubuntu|debian|linuxmint|pop)
            echo -e "${GREEN}[1/5]${NC} Updating system (apt)..."
            apt update -qq
            echo -e "${GREEN}[2/5]${NC} Installing PostgreSQL..."
            if command -v psql &> /dev/null; then
                echo "  PostgreSQL already installed"
            else
                apt install -y postgresql postgresql-contrib -qq
                systemctl enable postgresql
                systemctl start postgresql
                echo "  PostgreSQL installed and started"
            fi
            ;;
        arch|manjaro|endeavouros|garuda)
            echo -e "${GREEN}[1/5]${NC} Updating system (pacman)..."
            pacman -Sy --noconfirm
            echo -e "${GREEN}[2/5]${NC} Installing PostgreSQL..."
            if command -v psql &> /dev/null; then
                echo "  PostgreSQL already installed"
            else
                pacman -S postgresql --noconfirm
                # Initialize PostgreSQL data directory on Arch
                if [ ! -d /var/lib/postgres/data ] || [ -z "$(ls -A /var/lib/postgres/data 2>/dev/null)" ]; then
                    echo "  Initializing PostgreSQL database..."
                    mkdir -p /var/lib/postgres/data
                    chown -R postgres:postgres /var/lib/postgres
                    su - postgres -c "initdb -D /var/lib/postgres/data"
                fi
                systemctl enable postgresql
                systemctl start postgresql
                echo "  PostgreSQL installed and started"
            fi
            ;;
        fedora|rhel|centos|rocky|almalinux)
            echo -e "${GREEN}[1/5]${NC} Updating system (dnf)..."
            dnf check-update -q || true
            echo -e "${GREEN}[2/5]${NC} Installing PostgreSQL..."
            if command -v psql &> /dev/null; then
                echo "  PostgreSQL already installed"
            else
                dnf install -y postgresql-server postgresql-contrib -q
                # Initialize PostgreSQL on RHEL-based
                if [ ! -f /var/lib/pgsql/data/PG_VERSION ]; then
                    postgresql-setup --initdb
                fi
                systemctl enable postgresql
                systemctl start postgresql
                echo "  PostgreSQL installed and started"
            fi
            ;;
        opensuse*|sles)
            echo -e "${GREEN}[1/5]${NC} Updating system (zypper)..."
            zypper refresh -q
            echo -e "${GREEN}[2/5]${NC} Installing PostgreSQL..."
            if command -v psql &> /dev/null; then
                echo "  PostgreSQL already installed"
            else
                zypper install -y postgresql-server postgresql -q
                systemctl enable postgresql
                systemctl start postgresql
                echo "  PostgreSQL installed and started"
            fi
            ;;
        *)
            echo -e "${RED}Unsupported distribution: $DISTRO${NC}"
            echo "Supported: Ubuntu, Debian, Arch, Manjaro, Fedora, RHEL, CentOS, openSUSE"
            echo ""
            echo "Manual installation required:"
            echo "  1. Install PostgreSQL"
            echo "  2. Create database: createdb cold_db"
            echo "  3. Run: ./server"
            exit 1
            ;;
    esac
}

# ============================================
# INSTALL POSTGRESQL
# ============================================
install_postgresql

# Wait for PostgreSQL to be ready
echo "  Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if sudo -u postgres psql -c "SELECT 1" &>/dev/null; then
        break
    fi
    sleep 1
done

# ============================================
# CREATE DATABASE AND CONFIGURE USERS
# ============================================
echo -e "${GREEN}[3/5]${NC} Configuring database..."

# macOS uses different commands (no sudo -u postgres needed)
if [ "$DISTRO" = "macos" ]; then
    # Set password for current user (default superuser on macOS)
    echo "  Setting postgres password..."
    psql -d postgres -c "ALTER USER $(whoami) PASSWORD 'SecurePostgresPassword123';" > /dev/null 2>&1 || true

    # Create cold_user
    echo "  Creating cold_user..."
    psql -d postgres -c "CREATE USER cold_user WITH PASSWORD 'SecurePostgresPassword123';" 2>/dev/null || \
        psql -d postgres -c "ALTER USER cold_user PASSWORD 'SecurePostgresPassword123';" 2>/dev/null

    # Create database if not exists
    if psql -lqt | cut -d \| -f 1 | grep -qw cold_db; then
        echo "  Database 'cold_db' already exists"
    else
        psql -d postgres -c "CREATE DATABASE cold_db;" > /dev/null
        echo "  Database 'cold_db' created"
    fi

    # Grant cold_user access to cold_db
    psql -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE cold_db TO cold_user;" > /dev/null 2>&1

    # Configure pg_hba.conf for password authentication
    echo "  Configuring authentication..."
    PG_HBA=$(psql -d postgres -t -c "SHOW hba_file;" 2>/dev/null | xargs)
else
    # Linux: use sudo -u postgres
    # Set password for postgres user
    echo "  Setting postgres password..."
    sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'SecurePostgresPassword123';" > /dev/null 2>&1

    # Create cold_user (used in migrations for GRANT statements)
    echo "  Creating cold_user..."
    sudo -u postgres psql -c "CREATE USER cold_user WITH PASSWORD 'SecurePostgresPassword123';" 2>/dev/null || \
        sudo -u postgres psql -c "ALTER USER cold_user PASSWORD 'SecurePostgresPassword123';" 2>/dev/null

    # Create database if not exists
    if sudo -u postgres psql -lqt | cut -d \| -f 1 | grep -qw cold_db; then
        echo "  Database 'cold_db' already exists"
    else
        sudo -u postgres psql -c "CREATE DATABASE cold_db OWNER postgres;" > /dev/null
        echo "  Database 'cold_db' created"
    fi

    # Grant cold_user access to cold_db
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE cold_db TO cold_user;" > /dev/null 2>&1

    # Configure pg_hba.conf for password authentication
    echo "  Configuring authentication..."
    PG_HBA=$(sudo -u postgres psql -t -c "SHOW hba_file;" 2>/dev/null | xargs)
fi
if [ -f "$PG_HBA" ]; then
    # Check if already configured
    if ! grep -q "host.*cold_db.*127.0.0.1" "$PG_HBA" 2>/dev/null; then
        # Backup original
        cp "$PG_HBA" "${PG_HBA}.bak"

        # Add password auth entries at the beginning of the file (before other host entries)
        {
            echo "# Cold Storage authentication (added by installer)"
            echo "host    cold_db         all             127.0.0.1/32            scram-sha-256"
            echo "host    cold_db         all             ::1/128                 scram-sha-256"
            echo "host    cold_db         all             0.0.0.0/0               scram-sha-256"
            cat "$PG_HBA"
        } > "${PG_HBA}.new"
        mv "${PG_HBA}.new" "$PG_HBA"

        # Set ownership (different on macOS vs Linux)
        if [ "$DISTRO" = "macos" ]; then
            chown $(whoami) "$PG_HBA" 2>/dev/null || true
        else
            chown postgres:postgres "$PG_HBA"
        fi

        echo "  Password authentication configured"
    else
        echo "  Authentication already configured"
    fi

    # Enable listening on all interfaces (for remote DR access)
    if [ "$DISTRO" = "macos" ]; then
        PG_CONF=$(psql -d postgres -t -c "SHOW config_file;" 2>/dev/null | xargs)
    else
        PG_CONF=$(sudo -u postgres psql -t -c "SHOW config_file;" 2>/dev/null | xargs)
    fi

    if [ -f "$PG_CONF" ]; then
        if ! grep -q "^listen_addresses.*=.*'\*'" "$PG_CONF" 2>/dev/null; then
            # Update listen_addresses (use gsed on macOS if available, otherwise sed)
            if [ "$DISTRO" = "macos" ]; then
                if command -v gsed &> /dev/null; then
                    gsed -i "s/^#*listen_addresses.*/listen_addresses = '*'/" "$PG_CONF"
                else
                    sed -i '' "s/^#*listen_addresses.*/listen_addresses = '*'/" "$PG_CONF" 2>/dev/null || \
                        echo "listen_addresses = '*'" >> "$PG_CONF"
                fi
            else
                if grep -q "^#*listen_addresses" "$PG_CONF"; then
                    sed -i "s/^#*listen_addresses.*/listen_addresses = '*'/" "$PG_CONF"
                else
                    echo "listen_addresses = '*'" >> "$PG_CONF"
                fi
            fi
            echo "  Configured to listen on all interfaces"
        fi
    fi

    # Reload PostgreSQL to apply changes
    if [ "$DISTRO" = "macos" ]; then
        brew services restart postgresql@16 2>/dev/null || brew services restart postgresql 2>/dev/null || true
    else
        systemctl reload postgresql 2>/dev/null || systemctl restart postgresql
    fi
fi

echo "  Database configuration complete"

# ============================================
# SETUP SERVICE
# ============================================
echo -e "${GREEN}[4/5]${NC} Setting up service..."

if [ "$DISTRO" = "macos" ]; then
    # macOS: Use launchd
    INSTALL_DIR="/usr/local/cold-backend"

    # Stop existing service if running
    launchctl unload /Library/LaunchDaemons/com.cold.backend.plist 2>/dev/null || true
    pkill -f "cold-backend/server" 2>/dev/null || true
    sleep 1

    # Copy binary
    mkdir -p "$INSTALL_DIR"
    cp "$BINARY" "$INSTALL_DIR/server"
    chmod +x "$INSTALL_DIR/server"
else
    # Linux: Use systemd
    # Stop existing service if running (to avoid "Text file busy" error)
    if systemctl is-active --quiet cold-backend 2>/dev/null; then
        echo "  Stopping existing service..."
        systemctl stop cold-backend
        sleep 2
    fi

    # Copy binary to /opt
    mkdir -p /opt/cold-backend
    cp "$BINARY" /opt/cold-backend/server
    chmod +x /opt/cold-backend/server
fi

if [ "$DISTRO" = "macos" ]; then
    # macOS: Create launchd plist
    cat > /Library/LaunchDaemons/com.cold.backend.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.cold.backend</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/cold-backend/server</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/usr/local/cold-backend</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>DB_HOST</key>
        <string>localhost</string>
        <key>DB_PORT</key>
        <string>5432</string>
        <key>DB_USER</key>
        <string>cold_user</string>
        <key>DB_PASSWORD</key>
        <string>SecurePostgresPassword123</string>
        <key>DB_NAME</key>
        <string>cold_db</string>
        <key>REDIS_SERVICE_HOST</key>
        <string>192.168.15.210</string>
        <key>REDIS_SERVICE_PORT</key>
        <string>6379</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/usr/local/cold-backend/server.log</string>
    <key>StandardErrorPath</key>
    <string>/usr/local/cold-backend/server.log</string>
</dict>
</plist>
EOF

    # ============================================
    # START SERVER (macOS)
    # ============================================
    echo -e "${GREEN}[5/5]${NC} Starting server..."
    launchctl load /Library/LaunchDaemons/com.cold.backend.plist

    # Wait for server to start
    sleep 5

    # Check if running
    if pgrep -f "cold-backend/server" > /dev/null; then
        IP_ADDR=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}' | head -1)
        echo ""
        echo "╔════════════════════════════════════════════════════════════╗"
        echo "║                    INSTALLATION COMPLETE                    ║"
        echo "╠════════════════════════════════════════════════════════════╣"
        echo "║                                                            ║"
        echo "║  Server running on: http://${IP_ADDR:-localhost}:8080"
        echo "║                                                            ║"
        echo "║  Commands:                                                 ║"
        echo "║    View logs:    tail -f /usr/local/cold-backend/server.log"
        echo "║    Stop:         sudo launchctl unload /Library/LaunchDaemons/com.cold.backend.plist"
        echo "║    Start:        sudo launchctl load /Library/LaunchDaemons/com.cold.backend.plist"
        echo "║                                                            ║"
        echo "╚════════════════════════════════════════════════════════════╝"
        echo ""

        # Show recent logs
        echo "Recent logs:"
        tail -20 /usr/local/cold-backend/server.log 2>/dev/null || echo "  (waiting for logs...)"
    else
        echo -e "${RED}Server failed to start. Check logs:${NC}"
        tail -50 /usr/local/cold-backend/server.log 2>/dev/null
        exit 1
    fi
else
    # Linux: Copy templates if they exist
    if [ -d "$SCRIPT_DIR/templates" ]; then
        cp -r "$SCRIPT_DIR/templates" /opt/cold-backend/
        echo "  Templates copied"
    fi

    # Copy static files if they exist
    if [ -d "$SCRIPT_DIR/static" ]; then
        cp -r "$SCRIPT_DIR/static" /opt/cold-backend/
        echo "  Static files copied"
    fi

    # Copy migrations if they exist
    if [ -d "$SCRIPT_DIR/migrations" ]; then
        cp -r "$SCRIPT_DIR/migrations" /opt/cold-backend/
        echo "  Migrations copied"
    fi

    # Create systemd service
    cat > /etc/systemd/system/cold-backend.service << 'EOF'
[Unit]
Description=Cold Storage Backend
After=postgresql.service network.target
Wants=postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/cold-backend
ExecStart=/opt/cold-backend/server
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Environment - matches production DR setup
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_USER=cold_user"
Environment="DB_PASSWORD=SecurePostgresPassword123"
Environment="DB_NAME=cold_db"
Environment="REDIS_SERVICE_HOST=192.168.15.210"
Environment="REDIS_SERVICE_PORT=6379"

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable cold-backend

    # ============================================
    # START SERVER (Linux)
    # ============================================
    echo -e "${GREEN}[5/5]${NC} Starting server..."
    systemctl start cold-backend

    # Wait for server to start
    sleep 10

    # Check if running
    if systemctl is-active --quiet cold-backend; then
        IP_ADDR=$(ip -4 addr show | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | grep -v '127.0.0.1' | head -1 2>/dev/null || echo "localhost")
        echo ""
        echo "╔════════════════════════════════════════════════════════════╗"
        echo "║                    INSTALLATION COMPLETE                    ║"
        echo "╠════════════════════════════════════════════════════════════╣"
        echo "║                                                            ║"
        echo "║  Server running on: http://$IP_ADDR:8080"
        echo "║                                                            ║"
        echo "║  Commands:                                                 ║"
        echo "║    View logs:    journalctl -u cold-backend -f            ║"
        echo "║    Stop:         systemctl stop cold-backend              ║"
        echo "║    Start:        systemctl start cold-backend             ║"
        echo "║    Status:       systemctl status cold-backend            ║"
        echo "║                                                            ║"
        echo "╚════════════════════════════════════════════════════════════╝"
        echo ""

        # Show recent logs
        echo "Recent logs:"
        journalctl -u cold-backend -n 20 --no-pager
    else
        echo -e "${RED}Server failed to start. Check logs:${NC}"
        journalctl -u cold-backend -n 50 --no-pager
        exit 1
    fi
fi
