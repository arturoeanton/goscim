# Contributing to GoSCIM 🚀

First off, thanks for taking the time to contribute! ❤️

GoSCIM is a community-driven project, and we welcome contributions of all kinds. Whether you're fixing a bug, adding a feature, improving documentation, or helping other users, your contribution is valuable!

## 🌟 Ways to Contribute

### 🐛 Reporting Bugs
- Use the [Bug Report template](https://github.com/arturoeanton/goscim/issues/new?template=bug_report.md)
- Search existing issues first to avoid duplicates
- Provide detailed steps to reproduce
- Include system information and logs

### 💡 Suggesting Features
- Use the [Feature Request template](https://github.com/arturoeanton/goscim/issues/new?template=feature_request.md)
- Explain the use case and why it's needed
- Consider if it fits GoSCIM's scope and philosophy

### 📝 Improving Documentation
- Fix typos, unclear explanations, or missing information
- Add examples and tutorials
- Translate documentation to other languages
- Create video tutorials or blog posts

### 🔧 Contributing Code
- Fix bugs
- Implement new features
- Improve performance
- Add tests
- Refactor code

## 🚀 Quick Start for Contributors

### 1. Fork and Clone
```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/goscim.git
cd goscim
git remote add upstream https://github.com/arturoeanton/goscim.git
```

### 2. Set Up Development Environment
```bash
# Install dependencies
go mod tidy

# Start development database
docker-compose -f docker-compose.dev.yml up -d

# Run tests to ensure everything works
go test ./...
```

### 3. Create a Branch
```bash
git checkout -b feature/awesome-new-feature
# or
git checkout -b fix/critical-bug
```

## 📋 Development Guidelines

### Code Style
- Follow standard Go conventions (`gofmt`, `golint`)
- Use meaningful variable and function names
- Add comments for public functions and complex logic
- Keep functions small and focused

### Testing
- Write tests for new functionality
- Ensure existing tests pass
- Aim for good test coverage
- Include both unit and integration tests

### Documentation
- Update relevant documentation for new features
- Add examples in code comments
- Update API documentation if needed

## 🔄 Pull Request Process

### 1. Before You Submit
- [ ] Code follows the style guidelines
- [ ] Tests are written and passing
- [ ] Documentation is updated
- [ ] Commit messages are clear and descriptive

### 2. Creating the PR
- Use a descriptive title
- Fill out the PR template completely
- Reference any related issues
- Add screenshots if applicable

### 3. Review Process
- Maintainers will review your PR
- Address any feedback promptly
- Keep the conversation constructive
- Be patient - reviews take time!

## 🎯 Good First Issues

Looking for a place to start? Check out issues labeled:
- `good first issue` - Perfect for newcomers
- `help wanted` - We need community help
- `documentation` - Great for non-code contributions
- `bug` - Fix existing problems

## 🏆 Recognition

We appreciate all contributions! Contributors are recognized in:
- The project README
- Release notes for significant contributions
- GitHub contributor statistics
- Special mentions in community updates

## 💬 Getting Help

- 💭 [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions) - General questions and ideas
- 🐛 [GitHub Issues](https://github.com/arturoeanton/goscim/issues) - Bug reports and feature requests
- 📚 [Documentation](docs/) - Comprehensive guides and references

## 🌐 Community Guidelines

### Be Respectful
- Use welcoming and inclusive language
- Respect different viewpoints and experiences
- Accept constructive criticism gracefully
- Focus on what's best for the community

### Be Helpful
- Help newcomers and answer questions
- Share knowledge and best practices
- Provide constructive feedback
- Celebrate others' contributions

### Be Professional
- Keep discussions on-topic
- Avoid personal attacks or harassment
- Follow the [GitHub Community Guidelines](https://docs.github.com/en/github/site-policy/github-community-guidelines)

## 🛠️ Development Setup Details

### Prerequisites
- Go 1.16 or later
- Docker and Docker Compose
- Git

### Environment Setup
```bash
# Set up environment variables
cp .env.example .env
# Edit .env with your settings

# Install development tools
make setup-dev

# Start development services
make dev
```

### Testing
```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific tests
go test ./scim/parser -v

# Run integration tests
make test-integration
```

### Building
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build Docker image
make docker-build
```

## 📁 Project Structure

```
goscim/
├── main.go              # Entry point
├── scim/                # Core SCIM implementation
│   ├── config.go       # Configuration management
│   ├── types.go        # Data structures
│   ├── op_*.go         # SCIM operations
│   └── parser/         # Filter parser
├── config/             # Configuration files
│   ├── schemas/        # SCIM schemas
│   └── resourceType/   # Resource definitions
├── docs/               # Documentation
├── test/               # Test files
└── scripts/            # Build and utility scripts
```

## 🔐 Security

- Report security vulnerabilities privately to the maintainers
- Don't open public issues for security problems
- Follow responsible disclosure practices

## 📄 License

By contributing to GoSCIM, you agree that your contributions will be licensed under the MIT License.

## 🎉 Thank You!

Your contributions make GoSCIM better for everyone. Thank you for being part of our community!

---

Questions? Feel free to reach out in [Discussions](https://github.com/arturoeanton/goscim/discussions) or open an issue. We're here to help! 😊