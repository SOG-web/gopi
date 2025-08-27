#!/bin/bash

# GoPi Backend Test Runner
# This script provides convenient commands to run the test suite

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Change to project root
cd "$PROJECT_ROOT"

echo -e "${BLUE}🚀 GoPi Backend Test Runner${NC}"
echo -e "${BLUE}================================${NC}"

# Function to run all tests
run_all_tests() {
    echo -e "\n${YELLOW}Running all tests...${NC}"
    go test ./tests/... -v
}

# Function to run tests with coverage
run_tests_with_coverage() {
    echo -e "\n${YELLOW}Running tests with coverage...${NC}"
    go test ./tests/... -coverprofile=coverage.out
    echo -e "${GREEN}Coverage report generated: coverage.out${NC}"
    echo -e "${BLUE}To view coverage in browser: go tool cover -html=coverage.out${NC}"
}

# Function to run specific test suite
run_chat_tests() {
    echo -e "\n${YELLOW}Running chat module tests...${NC}"
    go test ./tests/chat -v
}

# Function to run chat handler tests
run_handler_tests() {
    echo -e "\n${YELLOW}Running chat handler tests...${NC}"
    go test ./tests/chat -run TestChatHandler -v
}

# Function to run chat service tests
run_service_tests() {
    echo -e "\n${YELLOW}Running chat service tests...${NC}"
    go test ./tests/chat -run TestChatService -v
}

# Function to run repository tests
run_repo_tests() {
    echo -e "\n${YELLOW}Running chat repository tests...${NC}"
    go test ./tests/chat -run TestGorm -v
}

# Function to run WebSocket tests
run_websocket_tests() {
    echo -e "\n${YELLOW}Running WebSocket tests...${NC}"
    go test ./tests/chat -run TestWebSocket -v
}

# Function to run tests in parallel
run_parallel_tests() {
    echo -e "\n${YELLOW}Running tests in parallel...${NC}"
    go test ./tests/... -parallel=4 -v
}

# Function to run tests with race detection
run_race_tests() {
    echo -e "\n${YELLOW}Running tests with race detection...${NC}"
    go test ./tests/... -race -v
}

# Function to show coverage summary
show_coverage() {
    if [ ! -f "coverage.out" ]; then
        echo -e "${RED}Coverage file not found. Run tests with coverage first.${NC}"
        exit 1
    fi
    echo -e "\n${YELLOW}Coverage Summary:${NC}"
    go tool cover -func=coverage.out
}

# Function to clean test artifacts
clean() {
    echo -e "\n${YELLOW}Cleaning test artifacts...${NC}"
    rm -f coverage.out
    echo -e "${GREEN}Cleaned!${NC}"
}

# Function to show help
show_help() {
    echo -e "\n${GREEN}Usage: $0 [COMMAND]${NC}"
    echo -e "\n${BLUE}Commands:${NC}"
    echo -e "  ${YELLOW}all${NC}          Run all tests"
    echo -e "  ${YELLOW}coverage${NC}     Run tests with coverage report"
    echo -e "  ${YELLOW}chat${NC}         Run chat module tests only"
    echo -e "  ${YELLOW}handlers${NC}     Run chat handler tests only"
    echo -e "  ${YELLOW}services${NC}     Run chat service tests only"
    echo -e "  ${YELLOW}repos${NC}        Run repository tests only"
    echo -e "  ${YELLOW}websockets${NC}   Run WebSocket tests only"
    echo -e "  ${YELLOW}parallel${NC}     Run tests in parallel"
    echo -e "  ${YELLOW}race${NC}         Run tests with race detection"
    echo -e "  ${YELLOW}summary${NC}      Show coverage summary"
    echo -e "  ${YELLOW}clean${NC}        Clean test artifacts"
    echo -e "  ${YELLOW}help${NC}         Show this help message"
    echo -e "\n${BLUE}Examples:${NC}"
    echo -e "  $0 all"
    echo -e "  $0 coverage"
    echo -e "  $0 handlers"
    echo -e "  $0 summary"
}

# Main logic
case "${1:-all}" in
    "all")
        run_all_tests
        ;;
    "coverage")
        run_tests_with_coverage
        ;;
    "chat")
        run_chat_tests
        ;;
    "handlers")
        run_handler_tests
        ;;
    "services")
        run_service_tests
        ;;
    "repos")
        run_repo_tests
        ;;
    "websockets")
        run_websocket_tests
        ;;
    "parallel")
        run_parallel_tests
        ;;
    "race")
        run_race_tests
        ;;
    "summary")
        show_coverage
        ;;
    "clean")
        clean
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        show_help
        exit 1
        ;;
esac

echo -e "\n${GREEN}✓ Test execution completed!${NC}"
