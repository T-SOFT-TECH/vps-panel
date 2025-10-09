# GitHub Repository Selector UI - Vercel Style

## ✅ What You'll See (Just Like Vercel!)

### When GitHub is Connected

When you visit the **New Project** page after connecting GitHub:

1. **Repository Picker Screen**
   - Shows ALL your GitHub repositories in a searchable list
   - Search box at the top to filter repos
   - Each repo shows:
     - GitHub icon
     - Repository name
     - Full name (username/repo-name)
     - 🔒 "Private" badge for private repos
     - Arrow to select →

2. **Search & Filter**
   - Type to instantly filter repos
   - Searches both repo name and full name
   - Works with hundreds of repos

3. **Select Repository**
   - Click any repo to select it
   - Automatically fills:
     - Repository URL
     - Default branch
     - Project name

4. **Configuration Form**
   - After selecting, shows:
     - Selected repo card with GitHub icon
     - "Change" button to go back
     - Auto-detects framework (SvelteKit, React, Next.js, etc.)
     - Pre-fills build settings
     - Ready to deploy!

### When GitHub is NOT Connected

Shows the regular form with:
- Manual Git URL entry
- Option to check "Private Repository"
- Manual credentials input

## UI Flow

### Flow 1: Connected GitHub User
```
1. New Project Page
   ↓
2. Shows: "Import Git Repository"
   - Search box
   - List of all repos
   - "Or enter manually →" button
   ↓
3. User clicks a repo
   ↓
4. Shows: Configuration form
   - Selected repo badge at top
   - All fields pre-filled
   - Framework auto-detected
   ↓
5. Click "Create Project"
   ↓
6. Deployed! (No credentials needed)
```

### Flow 2: Manual Entry
```
1. New Project Page
   ↓
2. User clicks "Or enter manually →"
   ↓
3. Shows: Traditional form
   - Git URL field
   - Optional: Private repo checkbox
   - Manual credentials
   ↓
4. Fill in details manually
   ↓
5. Create Project
```

## Visual Features

### Repository List Item
```
┌─────────────────────────────────────────────┐
│ 🔷 my-awesome-app     🔒 Private            │
│ username/my-awesome-app                  →  │
└─────────────────────────────────────────────┘
```

### Selected Repository Badge
```
┌─────────────────────────────────────────────┐
│ 🔷 username/my-awesome-app                  │
│ main branch                       [Change]  │
└─────────────────────────────────────────────┘
```

### Empty State
```
┌─────────────────────────────────────────────┐
│                    📁                        │
│         No repositories found               │
│                                             │
└─────────────────────────────────────────────┘
```

### Loading State
```
┌─────────────────────────────────────────────┐
│                    ⏳                        │
│         Loading repositories...             │
│                                             │
└─────────────────────────────────────────────┘
```

## Features Implemented

✅ **Auto-load repos** - Fetches on page mount if GitHub connected
✅ **Real-time search** - Filters as you type
✅ **Private repo badges** - Shows 🔒 for private repos
✅ **Smooth animations** - Fade in/out transitions
✅ **Loading states** - Spinner while fetching
✅ **Empty states** - Helpful messages when no repos
✅ **Repository selection** - Click to select
✅ **Auto-fill form** - Pre-fills all details
✅ **Framework detection** - Automatically detects on selection
✅ **Change repo** - Easy to go back and pick another
✅ **Manual override** - Can still enter manually
✅ **Responsive** - Works on mobile and desktop

## Interactions

### Hover Effects
- Repository items: Border changes to green
- Repository items: Background darkens slightly
- Buttons: Standard hover states

### Click Actions
1. **Select Repository** → Hides picker, shows form with repo info
2. **Change Repository** → Shows picker again, clears form
3. **Or enter manually** → Hides picker, shows empty form
4. **Search field** → Filters list in real-time

## Benefits vs Manual Entry

| Feature | GitHub OAuth | Manual Entry |
|---------|-------------|--------------|
| List all repos | ✅ Yes | ❌ No |
| Auto-fill URL | ✅ Yes | ❌ Manual |
| Auto-fill branch | ✅ Yes | ❌ Manual |
| Private repos | ✅ Works seamlessly | ⚠️ Need token |
| Framework detection | ✅ Automatic | ✅ Manual trigger |
| No credentials | ✅ Stored securely | ❌ Must enter |
| Search repos | ✅ Yes | ❌ N/A |

## Next Steps

To see this in action:

1. **Install OAuth packages**:
   ```bash
   cd backend
   go get golang.org/x/oauth2
   go get golang.org/x/oauth2/github
   ```

2. **Set up GitHub OAuth App** (see OAUTH_SETUP.md)

3. **Configure environment** with Client ID and Secret

4. **Build & run backend**:
   ```bash
   go build -o vps-panel.exe ./cmd/server
   ./vps-panel.exe
   ```

5. **Visit** http://localhost:5173/settings

6. **Click "Connect GitHub"**

7. **Go to New Project** → See all your repos! 🎉

## Screenshots Flow

### Step 1: Repository Selector
Shows search box and scrollable list of repos

### Step 2: Selected Repository
Shows green badge with repo info and "Change" button

### Step 3: Configuration Form
Shows pre-filled form ready to deploy

This is the exact same flow as:
- ✅ Vercel
- ✅ Netlify
- ✅ Railway
- ✅ Render

No more copying Git URLs manually! 🚀
