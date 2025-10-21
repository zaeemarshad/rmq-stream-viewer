#!/bin/bash

# Setup script to create test vhosts and configure permissions
# This needs to run before the publisher

HOST=${1:-localhost}
PORT=${2:-15672}
USER=${3:-root}
PASS=${4:-magic}

BASE_URL="http://$HOST:$PORT/api"

echo "Setting up test vhosts on $HOST:$PORT..."

# Create vhosts
for vhost in "test-vhost-1" "test-vhost-2" "test-vhost-3"; do
    echo "Creating vhost: $vhost"
    curl -s -u $USER:$PASS -X PUT "$BASE_URL/vhosts/$vhost" \
        -H "Content-Type: application/json" \
        -d '{"description":"Test vhost"}' || true
    
    # Set permissions for the user
    echo "Setting permissions for $USER on $vhost"
    curl -s -u $USER:$PASS -X PUT "$BASE_URL/permissions/$vhost/$USER" \
        -H "Content-Type: application/json" \
        -d '{"configure":".*","write":".*","read":".*"}' || true
done

echo "Vhost setup complete!"

