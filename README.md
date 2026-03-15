# PBX: AI-Powered Unified Communications Bridge

A comprehensive communication middleware designed to integrate telephony and messaging platforms with Large Language Model (LLM) intelligence. By bridging **Asterisk**, **WhatsApp**, and **Telegram**, the platform delivers interactive, AI-driven experiences for both voice and text interactions. Built upon the **Tibula RDBMS** framework, PBX provides granular control over user authentication, system-wide prompt engineering, and multi-language communication workflows.

## Key Capabilities

*   **Unified Multi-Platform Integration:** Centralized management of voice and messaging traffic across **Asterisk (AGI/ARI)**, **WhatsApp**, and **Telegram** endpoints.
*   **Advanced Conversational AI:**
    *   **LLM Interoperability:** Support for OpenAI-compatible APIs with **Model Context Protocol (MCP)** integration, allowing AI models to leverage custom internal tools.
    *   **Context Management:** Sophisticated stateful chat history tracking with automated session timeout handling.
    *   **Command Logic:** A robust system for parsing custom commands (e.g., `/reset`, `/mail`, `/ntfy`) to trigger asynchronous system actions.
*   **Media Processing Engine:**
    *   **Intelligent ASR/TTS:** Seamless integration with OpenAI (Whisper/TTS) or Google Speech services for high-fidelity audio conversion.
    *   **VoIP Optimization:** Real-time audio transcoding and Voice Activity Detection (VAD) specifically engineered for stable Asterisk performance.
*   **Scalable Architecture:** A modular plugin-based design facilitating custom chat actions and a database-driven configuration system for rapid deployment and maintenance.

## System Prerequisites

*   **Dependencies:** `ffmpeg` and `ffprobe` (for real-time media transcoding and metadata analysis).
*   **Database Management:** A RDBMS compatible with the Tibula framework.
*   **API Infrastructure:**
    *   **LLM Provider:** API access to OpenAI or an OpenAI-compatible endpoint.
    *   **Speech Services:** Credentials for OpenAI (Whisper/TTS) or Google Cloud Speech APIs.
    *   **Communication APIs:** Valid credentials for the Telegram Bot API or Meta/WhatsApp Business API.
*   **Telephony Environment:** A configured Asterisk server with PJSIP capability.

## Quick Start Guide

### 1. Installation
Clone the repository and compile the binary:

```bash
git clone https://github.com/eja/pbx
cd pbx
make
```

### 2. Initial Configuration
Execute the integrated setup wizard to configure database parameters, API tokens, and platform integration endpoints:

```bash
./build/pbx --wizard
```

### 3. Service Deployment
Initiate the service to deploy the web interface, the telephony AGI server, and the necessary webhook listeners:

```bash
./build/pbx --start
```
