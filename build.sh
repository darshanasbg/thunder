#!/bin/bash
# ----------------------------------------------------------------------------
# Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
#
# WSO2 LLC. licenses this file to you under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied. See the License for the
# specific language governing permissions and limitations
# under the License.
# ----------------------------------------------------------------------------

set -e

VERSION_FILE=version.txt
BINARY_NAME=thunder
VERSION=$(cat "$VERSION_FILE")
PRODUCT_FOLDER=${BINARY_NAME}-${VERSION}
# Server ports
FRONTEND_PORT=9090
BACKEND_PORT=8090

# Directories
OUTPUT_DIR=target
BUILD_DIR=$OUTPUT_DIR/.build
BACKEND_BASE_DIR=backend
BACKEND_DIR=$BACKEND_BASE_DIR/cmd/server
REPOSITORY_DIR=$BACKEND_BASE_DIR/cmd/server/repository
SERVER_SCRIPTS_DIR=$BACKEND_BASE_DIR/scripts
SERVER_DB_SCRIPTS_DIR=$BACKEND_BASE_DIR/dbscripts
SECURITY_DIR=repository/resources/security

function clean() {
    echo "Cleaning build artifacts..."
    rm -rf "$OUTPUT_DIR"
}

function build_backend() {
    echo "Building Go backend..."
    mkdir -p "$BUILD_DIR"
    go build -C "$BACKEND_BASE_DIR" -o "../$BUILD_DIR/$BINARY_NAME" ./cmd/server
}

function package() {
    echo "Packaging artifacts..."
    mkdir -p "$OUTPUT_DIR/$PRODUCT_FOLDER"
    cp "$BUILD_DIR/$BINARY_NAME" "$OUTPUT_DIR/$PRODUCT_FOLDER/"
    cp -r "$REPOSITORY_DIR" "$OUTPUT_DIR/$PRODUCT_FOLDER/"
    cp "$VERSION_FILE" "$OUTPUT_DIR/$PRODUCT_FOLDER/"
    cp -r "$SERVER_SCRIPTS_DIR" "$OUTPUT_DIR/$PRODUCT_FOLDER/"
    cp -r "$SERVER_DB_SCRIPTS_DIR" "$OUTPUT_DIR/$PRODUCT_FOLDER/"
    mkdir -p "$OUTPUT_DIR/$PRODUCT_FOLDER/$SECURITY_DIR"

    echo "Generating SSL certificates..."
    OPENSSL_ERR=$(
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout "$OUTPUT_DIR/$PRODUCT_FOLDER/$SECURITY_DIR/server.key" \
            -out "$OUTPUT_DIR/$PRODUCT_FOLDER/$SECURITY_DIR/server.crt" \
            -subj "/O=WSO2/OU=Thunder/CN=localhost" \
            > /dev/null 2>&1
    )
    if [[ $? -ne 0 ]]; then
        echo "Error generating SSL certificates: $OPENSSL_ERR"
        exit 1
    fi
    echo "Certificates generated successfully."
    
    echo "Creating zip file..."
    (cd "$OUTPUT_DIR" && zip -r "$PRODUCT_FOLDER.zip" "$PRODUCT_FOLDER")
    rm -rf "$OUTPUT_DIR/$PRODUCT_FOLDER" "$BUILD_DIR"
}

function test_integration() {
    echo "Running integration tests..."
    go run -C ./tests/integration ./main.go
}

function run() {
    echo "=== Cleaning build output ==="
    clean

    echo "=== Building backend ==="
    build_backend

    echo "=== Packaging artifacts ==="
    package

    echo "=== Ensuring server certificates exist ==="
    if [[ ! -f "$BACKEND_DIR/$SECURITY_DIR/server.crt" || ! -f "$BACKEND_DIR/$SECURITY_DIR/server.key" ]]; then
        mkdir -p "$BACKEND_DIR/$SECURITY_DIR"
        echo "Generating SSL certificates..."
        OPENSSL_ERR=$(
            openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
                -keyout "$BACKEND_DIR/$SECURITY_DIR/server.key" \
                -out "$BACKEND_DIR/$SECURITY_DIR/server.crt" \
                -subj "/O=WSO2/OU=Thunder/CN=localhost" \
                > /dev/null 2>&1
        )
        if [[ $? -ne 0 ]]; then
            echo "Error generating SSL certificates: $OPENSSL_ERR"
            exit 1
        fi
        echo "Certificates generated successfully."
    else
        echo "Certificates already exist."
    fi

    # Kill known ports
    function kill_port() {
        local port=$1
        lsof -ti tcp:$port | xargs kill -9 2>/dev/null || true
    }

    kill_port $FRONTEND_PORT
    kill_port $BACKEND_PORT

    echo "=== Starting frontend on https://localhost:$FRONTEND_PORT ==="
    FRONTEND_PORT=$FRONTEND_PORT pnpm --filter gate start &
    FRONTEND_PID=$!

    echo "=== Starting backend on https://localhost:$BACKEND_PORT ==="
    BACKEND_PORT=$BACKEND_PORT go run -C "$BACKEND_DIR" . &
    BACKEND_PID=$!

    echo ""
    echo "🚀 Servers running:"
    echo "👉 Frontend: https://localhost:$FRONTEND_PORT"
    echo "👉 Backend : https://localhost:$BACKEND_PORT"
    echo "Press Ctrl+C to stop."

    trap 'echo -e "\nStopping servers..."; kill $FRONTEND_PID $BACKEND_PID; exit' SIGINT

    wait $FRONTEND_PID
    wait $BACKEND_PID
}

case "$1" in
    clean)
        clean
        ;;
    build)
        build_backend
        package
        ;;
    test)
        test_integration
        ;;
    run)
        run
        ;;
    *)
        echo "Usage: ./build.sh {clean|build|test|run}"
        exit 1
        ;;
esac
