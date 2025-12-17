#!/bin/bash

echo "========================================"
echo "GitHub Repository Setup Script"
echo "Cold Storage Management System"
echo "========================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Get GitHub username
echo -e "${BLUE}Enter your GitHub username:${NC}"
read -r GITHUB_USERNAME

if [ -z "$GITHUB_USERNAME" ]; then
    echo -e "${RED}Error: GitHub username is required${NC}"
    exit 1
fi

# Repository name
REPO_NAME="cold-storage-management-system"

echo ""
echo -e "${YELLOW}Repository Details:${NC}"
echo "  Username: $GITHUB_USERNAME"
echo "  Repository: $REPO_NAME"
echo "  Full URL: https://github.com/$GITHUB_USERNAME/$REPO_NAME.git"
echo ""

# Ask if repository is created
echo -e "${BLUE}Have you created the repository on GitHub? (y/n)${NC}"
echo "If not, go to: https://github.com/new"
echo "Repository name: $REPO_NAME"
echo "Make it private/public, and DO NOT initialize with README"
read -r CREATED

if [ "$CREATED" != "y" ]; then
    echo ""
    echo -e "${YELLOW}Please create the repository first:${NC}"
    echo "1. Go to: https://github.com/new"
    echo "2. Repository name: $REPO_NAME"
    echo "3. Choose private or public"
    echo "4. DO NOT check 'Initialize with README'"
    echo "5. Click 'Create repository'"
    echo ""
    echo "Then run this script again!"
    exit 0
fi

echo ""
echo -e "${GREEN}Setting up Git remote...${NC}"

# Remove existing remote if any
git remote remove origin 2>/dev/null

# Add new remote
REPO_URL="https://github.com/$GITHUB_USERNAME/$REPO_NAME.git"
git remote add origin "$REPO_URL"

echo -e "${GREEN}✓ Remote added successfully${NC}"
echo ""

# Show current status
echo -e "${BLUE}Current git status:${NC}"
git status
echo ""

# Ask to push
echo -e "${YELLOW}Ready to push to GitHub?${NC}"
echo "This will push your code to: $REPO_URL"
read -p "Continue? (y/n) " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo -e "${GREEN}Pushing to GitHub...${NC}"

    # Push to GitHub
    git push -u origin main

    if [ $? -eq 0 ]; then
        echo ""
        echo -e "${GREEN}========================================${NC}"
        echo -e "${GREEN}✓ Successfully pushed to GitHub!${NC}"
        echo -e "${GREEN}========================================${NC}"
        echo ""
        echo "Your repository is now available at:"
        echo -e "${BLUE}https://github.com/$GITHUB_USERNAME/$REPO_NAME${NC}"
        echo ""
        echo "Next steps:"
        echo "1. Visit your repository and verify files"
        echo "2. Add repository description and topics"
        echo "3. Share with your team!"
    else
        echo ""
        echo -e "${RED}========================================${NC}"
        echo -e "${RED}✗ Push failed${NC}"
        echo -e "${RED}========================================${NC}"
        echo ""
        echo -e "${YELLOW}Common issues:${NC}"
        echo "1. Authentication failed?"
        echo "   - GitHub requires Personal Access Token (not password)"
        echo "   - Create token at: https://github.com/settings/tokens"
        echo "   - Select 'repo' scope"
        echo "   - Use token as password when prompted"
        echo ""
        echo "2. Permission denied?"
        echo "   - Make sure repository exists on GitHub"
        echo "   - Verify repository name is correct"
        echo "   - Check if you have write access"
        echo ""
        echo "For detailed help, see: GITHUB_SETUP.md"
    fi
else
    echo ""
    echo "Push cancelled. You can push manually later with:"
    echo "  git push -u origin main"
fi

echo ""
echo -e "${BLUE}Git remote configured:${NC}"
git remote -v
