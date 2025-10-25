# vox: Your AI Co-pilot for Product Interviews
> 95% Vibe coded with Jules

## What is vox?
vox is a command-line tool that uses AI to help you practice for product interviews. Think of it as a friendly, programmable interviewer that can simulate everything from product sense and design discussions to behavioral and leadership questions. It's designed to be simple, extensible, and a little bit of fun.

## Why vox?
The goal of vox is to make product interview practice more accessible and less stressful. Landing a product role is tough, and practicing your interview skills is key. vox provides a safe, low-stakes way to get in the reps, helping you articulate your product vision, practice your responses, and build confidence before the real thing.

## Getting Started
Ready to give it a whirl? Here’s how to get up and running in a few simple steps.

### 1. Installation
As a Go CLI, you can install vox directly:
```bash
go install github.com/andrewhowdencom/vox@latest
```

### 2. Configuration
vox uses a `.vox.yaml` file to define different interview "topics". This is where you tell it what questions to ask or what kind of interviewer personality to adopt.

Create your config file at `~/.config/.vox.yaml` and add your first topics. Here’s an example to get you started:

```yaml
providers:
  gemini:
    # Pro-tip: you can use any model you want here!
    model: "gemini-flash-latest"

interviews:
    # A static interview with pre-written questions
    - id: behavioural-interview
      provider: static
      questions:
        - "Tell me about a time you had to deal with a difficult coworker."
        - "What is your greatest weakness?"
        - "How do you handle stress and pressure?"

    # An AI-powered interview using Gemini for product sense
    - id: product-sense-interview
      provider: gemini
      prompt: "You are an interviewer conducting a product sense interview for a Senior Product Manager position."
```

### 3. Run an Interview
Once your config is set up, you can start an interview from your terminal:

```bash
vox interview start --topic behavioural-interview
```

**Using the Gemini Provider**

If you're using the `gemini` provider, you'll need to include your API key. You can do this by passing the `--api-key` flag:

```bash
vox interview start --topic product-sense-interview --api-key YOUR_GEMINI_API_KEY
```

## Features
- **Multiple Providers**: Mix and match interview styles. Use the `static` provider for a predictable set of questions or `gemini` for dynamic, AI-powered conversations.
- **Slack Integration**: Conduct interviews directly within your Slack workspace! Just run the `/vox interview start --topic <your-topic>` command.
- **Extensible by Design**: Built with a hexagonal architecture, making it easy for developers to add new interview providers, UIs (want a web version?), or other fun features.

## Architecture
For those who like to peek under the hood, vox is built using a **Hexagonal Architecture** (also known as Ports and Adapters). In simple terms, this means the core logic of the application (the "domain") is completely decoupled from the outside world.

- **The Core**: The `internal/domain` package handles the interview logic.
- **Ports**: The `internal/ports` package contains the "entry points" to the application, like the CLI (`cobra`) and the Slack server.
- **Adapters**: The `internal/adapters` package holds the different implementations for things like question providers (`static`, `gemini`) and user interfaces (`terminal`, `slack`).

This structure keeps the code clean, testable, and super easy to extend.
