cmake_minimum_required(VERSION 3.30)
project(eiene VERSION 0.1)

set(GO_MODULE_PATH "${CMAKE_SOURCE_DIR}")

add_custom_target(
 build
 COMMAND go build -o "${CMAKE_BINARY_DIR}/${PROJECT_NAME}" "${GO_MODULE_PATH}/main.go"
 WORKING_DIRECTORY "${GO_MODULE_PATH}"
 COMMENT "Building ${PROJECT_NAME}"
)

add_custom_target(
 test
 COMMAND go test ./...
 WORKING_DIRECTORY "${GO_MODULE_PATH}"
 COMMENT "Running ${PROJECT_NAME} tests"
)

add_custom_target(
 test_verbose
 COMMAND go test -v ./...
 WORKING_DIRECTORY "${GO_MODULE_PATH}"
 COMMENT "Running ${PROJECT_NAME} tests - verbose"
)

add_custom_target(
 coverage
 COMMAND go test -coverprofile="${CMAKE_BINARY_DIR}/coverage.out" ./...
 COMMAND go tool cover -html="${CMAKE_BINARY_DIR}/coverage.out"
 WORKING_DIRECTORY "${GO_MODULE_PATH}"
 COMMENT "Running ${PROJECT_NAME} tests - coverage"
)

add_custom_target(
 clean_go
 COMMAND go clean
 COMMAND rm -rf "${CMAKE_BINARY_DIR}/${PROJECT_NAME}"
 WORKING_DIRECTORY "${GO_MODULE_PATH}"
 COMMENT "Cleaning ${PROJECT_NAME}"
)

add_dependencies(build clean_go)
add_dependencies(test build)
add_dependencies(coverage build)
