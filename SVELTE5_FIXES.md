# Svelte 5 Syntax Fixes

## Issue
Svelte 5 removed event modifiers like `|preventDefault`, `|stopPropagation`, etc.

## Error
```
'onsubmit|preventDefault' is not a valid attribute name
```

## Solution

### ❌ Old Svelte 4 Syntax
```svelte
<form onsubmit|preventDefault={handleSubmit}>
```

### ✅ New Svelte 5 Syntax
```svelte
<script>
  function handleSubmit(e: Event) {
    e.preventDefault();  // Handle preventDefault manually
    // Your code here
  }
</script>

<form onsubmit={handleSubmit}>
```

## Files Fixed

1. ✅ `routes/login/+page.svelte`
2. ✅ `routes/register/+page.svelte`
3. ✅ `routes/(app)/projects/new/+page.svelte`

## All Event Modifiers Removed

In Svelte 5, these modifiers are no longer supported:
- ❌ `|preventDefault`
- ❌ `|stopPropagation`
- ❌ `|passive`
- ❌ `|nonpassive`
- ❌ `|capture`
- ❌ `|once`
- ❌ `|self`
- ❌ `|trusted`

**Solution:** Handle them manually in your event handler function:

```typescript
function handleEvent(e: Event) {
  e.preventDefault();     // Instead of |preventDefault
  e.stopPropagation();    // Instead of |stopPropagation
  // etc.
}
```

## ✅ Fixed!

All forms now use proper Svelte 5 syntax. The frontend should compile without errors!
