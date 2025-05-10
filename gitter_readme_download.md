# Gitter - How to Use

Gitter is a simplified version control system that helps you track changes to your files and maintain a history of your project. This guide will show you exactly how to use each feature.

## Installation and Setup

### 1. Build the Application

```bash
# From your gitter project directory
go build -o gitter cmd/gitter/main.go
```

### 2. Verify Installation

```bash
./gitter help
```

## Basic Usage

### Starting a New Project

1. **Create a new directory for your project**:
   ```bash
   mkdir my-project
   cd my-project
   ```

2. **Initialize Gitter**:
   ```bash
   ../gitter init
   ```
   This creates a hidden `.gitter` directory to store version history.

### Tracking Your First Files

1. **Create some files**:
   ```bash
   echo "# My Project" > README.md
   echo "console.log('Hello!');" > app.js
   mkdir src
   echo "print('Hello')" > src/main.py
   ```

2. **Check what files are available**:
   ```bash
   ../gitter status
   ```
   You'll see all untracked files listed.

3. **Add files to staging area**:
   ```bash
   # Add specific files
   ../gitter add README.md
   ../gitter add app.js
   
   # Or add all files at once
   ../gitter add .
   ```

4. **Create your first commit**:
   ```bash
   ../gitter commit -m "Initial project setup"
   ```

## Detailed Command Usage

### 1. `init` - Initialize Repository

**What it does**: Creates a new Gitter repository in the current directory.

```bash
../gitter init
```

**When to use**: Once, when starting a new project.

### 2. `status` - Check Repository Status

**What it does**: Shows which files are tracked, modified, or ready to commit.

```bash
../gitter status
```

**Example output**:
```
Changes to be committed:
  modified: file1.txt

Changes not staged for commit:
  modified: file2.txt

Untracked files:
  newfile.txt
```

**When to use**: Before adding files or committing, to see what's changed.

### 3. `add` - Stage Files

**What it does**: Prepares files to be included in the next commit.

```bash
# Add a single file
../gitter add filename.txt

# Add multiple specific files
../gitter add file1.txt file2.txt file3.txt

# Add all JavaScript files
../gitter add *.js

# Add all files in a directory
../gitter add src/

# Add all files and directories
../gitter add .
```

**When to use**: After making changes and before committing.

### 4. `commit` - Save Changes

**What it does**: Records your staged changes with a message.

```bash
# Basic commit
../gitter commit -m "Your commit message here"

# Commit all modified files (doesn't include new files)
../gitter commit -am "Update existing files"
```

**Good commit messages**:
- "Add user authentication feature"
- "Fix bug in login function"
- "Update README with new instructions"

**When to use**: When you've made a logical set of changes you want to save.

### 5. `log` - View History

**What it does**: Shows a list of all previous commits.

```bash
../gitter log
```

**Example output**:
```
commit abc123...
Author: user
Date: Mon Jan 26 15:30:00 2025 +0530

    Add user authentication

commit def456...
Author: user
Date: Mon Jan 26 14:20:00 2025 +0530

    Initial commit
```

**When to use**: To see what changes have been made over time.

### 6. `diff` - See Changes

**What it does**: Shows exactly what has changed in your files.

```bash
# See all changes
../gitter diff

# See changes in a specific file
../gitter diff filename.txt

# See changes in a directory
../gitter diff src/
```

**Example output**:
```
--- a/app.js
+++ b/app.js
@@ -1,3 +1,4 @@
 console.log('Hello');
+console.log('New feature');
```

**When to use**: To review what you've changed before committing.

## Practical Workflows

### Workflow 1: Daily Development

```bash
# 1. Check what's changed
../gitter status

# 2. Review your changes
../gitter diff

# 3. Add specific files you want to commit
../gitter add src/feature.js
../gitter add README.md

# 4. Commit with a clear message
../gitter commit -m "Implement user dashboard feature"

# 5. View history
../gitter log
```

### Workflow 2: Working with Multiple Files

```bash
# Create multiple files
echo "Component 1" > component1.js
echo "Component 2" > component2.js
echo "Styles" > styles.css

# Add related files together
../gitter add component1.js component2.js styles.css
../gitter commit -m "Add new UI components"

# Or add by pattern
../gitter add *.js
../gitter commit -m "Add all JavaScript files"
```

### Workflow 3: Making Changes to Existing Project

```bash
# 1. Make changes to files
echo "Updated content" >> README.md

# 2. See what changed
../gitter diff README.md

# 3. Stage and commit
../gitter add README.md
../gitter commit -m "Update README with new features"
```

## Tips and Best Practices

### 1. Use Meaningful Commit Messages
- Good: "Fix null pointer exception in user login"
- Bad: "fix bug"

### 2. Check Status Frequently
```bash
../gitter status  # Run this often to stay aware of changes
```

### 3. Review Changes Before Committing
```bash
../gitter diff  # Always review what you're committing
```

### 4. Group Related Changes
```bash
# Group related changes in one commit
../gitter add auth/login.js auth/logout.js
../gitter commit -m "Complete authentication features"
```

### 5. Use Descriptive File Organization
```bash
# Organize files logically
mkdir -p src/{components,utils,styles}
../gitter add src/
```

## Common Scenarios

### Scenario 1: Starting a Web Project

```bash
# Initialize
../gitter init

# Create structure
mkdir -p src/js src/css images
echo "<html>My Site</html>" > index.html
echo "body { margin: 0; }" > src/css/styles.css
echo "console.log('Site loaded');" > src/js/app.js

# Add all files
../gitter add .
../gitter commit -m "Initial website structure"
```

### Scenario 2: Making Progressive Changes

```bash
# Day 1: Add header
echo "<header>Header</header>" >> index.html
../gitter add index.html
../gitter commit -m "Add header section"

# Day 2: Add navigation
echo "<nav>Navigation</nav>" >> index.html
../gitter add index.html
../gitter commit -m "Add navigation menu"

# Day 3: Add footer
echo "<footer>Footer</footer>" >> index.html
../gitter add index.html
../gitter commit -m "Add footer section"
```

### Scenario 3: Working with Large Projects

```bash
# Create complex structure
mkdir -p backend/src backend/tests frontend/src frontend/public

# Add backend files
echo "Backend code" > backend/src/server.js
../gitter add backend/
../gitter commit -m "Add backend structure"

# Add frontend files
echo "Frontend code" > frontend/src/app.js
../gitter add frontend/
../gitter commit -m "Add frontend structure"
```

## Troubleshooting

### Problem 1: "not a gitter repository"
**Solution**: You need to initialize first
```bash
../gitter init
```

### Problem 2: "nothing to commit"
**Solution**: Add files first
```bash
../gitter add yourfile.txt
```

### Problem 3: "required flag --message not set"
**Solution**: Always include a message
```bash
../gitter commit -m "Your message here"
```

### Problem 4: Can't add directory
**Solution**: Gitter automatically adds all files in a directory
```bash
../gitter add dirname/  # This adds all files in the directory
```

## Quick Reference

```bash
# Initialize
../gitter init

# Check status
../gitter status

# Add files
../gitter add filename.txt
../gitter add .

# Commit
../gitter commit -m "message"

# View history
../gitter log

# See changes
../gitter diff

# Get help
../gitter help
../gitter help commit
```

This covers everything you need to know to use Gitter effectively!
