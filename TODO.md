# Fake SSH Server Implementation Plan

## Development Stages

### 1. Project Setup
- [x] Initialize Go module
- [x] Create basic directory structure
- [x] Add necessary dependencies

### 2. Configuration Creation
- [x] Implement configuration structure
- [x] Add command-line flag support
- [x] Add configuration file support

### 3. Core Functionality Implementation
- [x] Set up SSH server
- [x] Implement authentication handler
- [x] Implement credential logging mechanism

### 4. OpenSSH Emulation
- [x] Configure greeting banner
- [x] Emulate OpenSSH behavior
- [x] Implement rejection of all authentication attempts

### 5. Logging
- [x] Implement client IP address logging
- [x] Implement username logging
- [x] Implement password logging
- [x] Format logs in analysis-friendly format

### 6. Testing
- [x] Write unit tests for configuration module
- [x] Write unit tests for SSH server module
- [x] Write unit tests for logging module
- [x] Integration testing

### 7. Building and Running
- [x] Add Makefile for project building
- [x] Create Dockerfile for containerization
- [x] Documentation for launching and usage

### 8. Additional Improvements
- [x] Integration of zerolog for structured logging
- [x] Support for logging to stdout for use with journald
- [x] Creation of Docker image based on Alpine Linux
- [x] Addition of systemd unit file for container launch
- [x] Documentation update with information about new features 

### 9. Extended SSH Server Functionality
- [x] Support for persistent SSH key (built-in, from file)
- [x] Option to generate new key or use existing key
- [x] Configurable SSH server version

### 10. Enhanced Logging 
- [x] Multiple log formats (JSON, pretty)
- [x] Configurable log destination (file, stdout)
- [x] Improved log structure with zerolog 