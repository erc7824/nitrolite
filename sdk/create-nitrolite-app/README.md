# create-nitrolite-app

A CLI tool to quickly create new Nitrolite applications with various templates and configurations.

## Usage

### Interactive Mode (Recommended)

```bash
npx create-nitrolite-app
```

### Quick Start

```bash
# Create with default template
npx create-nitrolite-app my-app

# Create with specific template
npx create-nitrolite-app my-app --template nextjs-app

# Skip prompts and use defaults
npx create-nitrolite-app my-app --yes
```

### Advanced Usage

```bash
# Create without git initialization
npx create-nitrolite-app my-app --no-git

# Create without installing dependencies
npx create-nitrolite-app my-app --no-install

# Combine options
npx create-nitrolite-app my-app --template nextjs-app --no-git --yes
```

## Available Templates

### `nextjs-app`

- Next.js 15 with App Router
- TypeScript and TailwindCSS
- Server-side rendering
- Optimized for production

## CLI Options

| Option              | Description                  | Default      |
| ------------------- | ---------------------------- | ------------ |
| `--template <name>` | Template to use              | `react-vite` |
| `--no-git`          | Skip git initialization      | `false`      |
| `--no-install`      | Skip dependency installation | `false`      |
| `--yes`             | Skip prompts, use defaults   | `false`      |
| `--help`            | Show help                    | -            |
| `--version`         | Show version                 | -            |

## Development

### Build

```bash
npm run build
```

### Development

```bash
npm run dev
```

### Test

```bash
npm test
```

## Project Structure

```
my-nitrolite-app/
├── src/
│   ├── components/        # React/Vue components
│   ├── hooks/             # Custom hooks (React)
│   ├── composables/       # Composables (Vue)
│   ├── utils/             # Utility functions
│   └── main.tsx           # Entry point
├── public/                # Static assets
├── package.json
└── README.md
```

## Next Steps

After creating your project:

1. **Navigate to your project**:

    ```bash
    cd my-nitrolite-app
    ```

2. **Install dependencies** (if not done automatically):

    ```bash
    npm install
    ```

3. **Start development server**:

    ```bash
    npm run dev
    ```

4. **Configure your WebSocket endpoint** in the config file

5. **Start building your Nitrolite application**!

## Documentation

- [Nitrolite SDK Documentation](https://github.com/erc7824/nitrolite)
- [Examples](../../examples/)
- [Issue Tracker](https://github.com/erc7824/nitrolite/issues)

## Contributing

Contributions are welcome! Please read the contributing guide in the main repository.

## License

ISC License - see the main repository for details.
