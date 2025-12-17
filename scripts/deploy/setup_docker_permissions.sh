#!/bin/bash

# Add current user to docker group
echo "Adding user to docker group..."
sudo usermod -aG docker $USER

echo ""
echo "✓ User added to docker group successfully!"
echo ""
echo "IMPORTANT: You need to apply the group change by running ONE of these:"
echo ""
echo "Option A (preferred): Logout and login again"
echo ""
echo "Option B (quick): Run this command in your terminal:"
echo "  newgrp docker"
echo ""
echo "After that, you can verify with:"
echo "  groups"
echo ""
echo "Then run the deployment script:"
echo "  ./deploy_monitoring.sh"
echo ""
