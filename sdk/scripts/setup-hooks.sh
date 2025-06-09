#!/bin/bash

# Setup Git hooks for automated validation
echo "🔧 Setting up Git hooks for automated SDK validation..."

# Create .git/hooks directory if it doesn't exist
mkdir -p .git/hooks

# Create pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

echo "🔍 Running pre-commit validation..."

# Change to SDK directory
cd "$(dirname "$0")/../../sdk" || exit 1

# Run TypeScript validation
echo "📝 Validating TypeScript types..."
npm run validate

if [ $? -ne 0 ]; then
    echo "❌ Pre-commit validation failed!"
    echo "   Please fix the issues above before committing."
    exit 1
fi

echo "✅ Pre-commit validation passed!"
EOF

# Make pre-commit hook executable
chmod +x .git/hooks/pre-commit

# Create post-merge hook to regenerate types after merges
cat > .git/hooks/post-merge << 'EOF'
#!/bin/bash

echo "🔄 Post-merge: Regenerating types..."

# Change to contract directory and build
cd "$(dirname "$0")/../../contract" || exit 1
if [ -f "foundry.toml" ]; then
    echo "📦 Building contracts..."
    forge build
fi

# Change to SDK directory and regenerate types
cd "../sdk" || exit 1
echo "🔧 Regenerating TypeScript types..."
npm run codegen

echo "✅ Post-merge setup complete!"
EOF

# Make post-merge hook executable
chmod +x .git/hooks/post-merge

echo "✅ Git hooks installed successfully!"
echo ""
echo "📋 Hooks installed:"
echo "   • pre-commit: Validates types before commits"
echo "   • post-merge: Regenerates types after merges"
echo ""
echo "🚀 Your SDK now has automated reliability checks!"
