#!/bin/bash

echo "========================================"
echo "   Reset Database for Testing"
echo "========================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Warning
echo -e "${RED}WARNING: This will DELETE ALL USER DATA!${NC}"
echo ""
echo "This will:"
echo "  - Delete all users (except admin)"
echo "  - Delete all customers"
echo "  - Delete all entries"
echo "  - Delete all room entries"
echo "  - Delete all payments"
echo "  - Reset all ID sequences to 1"
echo "  - Keep only default admin user (admin@cold.com / admin123)"
echo ""

read -p "Are you sure you want to continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Reset cancelled."
    exit 0
fi

echo ""
echo -e "${YELLOW}Resetting database...${NC}"

# Run the reset SQL script
if command -v docker &> /dev/null; then
    # Try with docker
    docker exec -i cold-storage-postgres psql -U postgres -d cold_db < migrations/reset_data.sql

    if [ $? -eq 0 ]; then
        echo ""
        echo -e "${GREEN}✓ Database reset successful!${NC}"
        echo ""
        echo "Default credentials:"
        echo "  Email: admin@cold.com"
        echo "  Password: admin123"
    else
        echo ""
        echo -e "${RED}✗ Reset failed!${NC}"
        echo "Make sure PostgreSQL container is running."
        exit 1
    fi
else
    # Try with psql directly
    psql -U postgres -d cold_db < migrations/reset_data.sql

    if [ $? -eq 0 ]; then
        echo ""
        echo -e "${GREEN}✓ Database reset successful!${NC}"
        echo ""
        echo "Default credentials:"
        echo "  Email: admin@cold.com"
        echo "  Password: admin123"
    else
        echo ""
        echo -e "${RED}✗ Reset failed!${NC}"
        exit 1
    fi
fi

echo ""
echo "Database is now ready for testing!"
