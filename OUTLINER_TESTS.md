# Outliner Frontend Tests

## Overview

This document describes the test suite for the Composter outliner frontend (`static/outliner.js`). The tests ensure that the keyboard-driven outliner works correctly and handles edge cases properly.

## Test Infrastructure

### Technology Stack
- **Testing Framework**: Jest 30.x
- **Test Environment**: jsdom (simulates browser DOM)
- **Test Files**: 
  - `static/outliner.test.js` - Main test suite
  - `static/outliner.module.js` - Module wrapper for loading the OutlineManager class

### Running Tests

```bash
# Run all tests once
npm test

# Run tests in watch mode (auto-rerun on file changes)
npm run test:watch

# Run tests with coverage report
npm run test:coverage
```

### Test Structure

The test suite is organized into the following categories:

## Test Categories

### 1. Basic Setup
Tests that the OutlineManager initializes correctly with and without initial content.

**Test Cases:**
- Initialize with empty content
- Initialize with provided HTML content

### 2. Indentation Level Calculation
Tests the core function that determines how many levels deep a line is indented.

**Test Cases:**
- Calculate indent level for no indentation (0 spaces)
- Calculate indent level for 2-space indent (1 level)
- Calculate indent level for 4-space indent (2 levels)
- Calculate indent level for 6-space indent (3 levels)
- Handle odd number of spaces (floor division)

### 3. Line and Children Indices
Tests the functions that find parent-child relationships in the outline hierarchy.

**Test Cases:**
- Find direct children of a parent line
- Find all descendants (children, grandchildren, etc.)
- Return empty array when no children exist

### 4. Content Conversion
Tests conversion between plain text (internal representation) and HTML (storage format).

**Test Cases:**
- Convert plain text with indentation to HTML with margin-left styling
- Convert HTML with margin-left styling to plain text with spaces
- Handle empty lines in conversion (should use `<br>` tags)

### 5. Indenting Operations
Tests the Tab and Shift+Tab functionality for changing indentation levels.

**Test Cases:**
- Indent a line without children
- Indent line with all its children (hierarchical indent)
- Unindent a line that has indentation
- Not unindent a line with no indentation (no-op)
- Unindent line with all its children (hierarchical unindent)

### 6. Enter Key - New Line Insertion
Tests the Enter key functionality for creating new lines with proper indentation.

**Test Cases:**
- Insert new line with same indentation at cursor position
- Insert new line with same indentation for indented line
- Insert new line in middle of text (split behavior)

### 7. Enter + Tab Bug Fix
**Critical bug tests** addressing the issue where pressing Enter then Tab would incorrectly indent the previous line instead of the new line.

**Test Cases:**
- After Enter, Tab should indent the **new** line, not the previous line
- After Enter on indented line, Tab should indent the new line correctly

**Expected Behavior:**
```
Before: "Line 1|"  (cursor at end, | = cursor)
After Enter: "Line 1\n|"  (new line, cursor on new line)
After Tab: "Line 1\n  |"  (new line is indented, Line 1 unchanged)
```

### 8. Moving Lines
Tests the Alt+Up and Alt+Down functionality for reordering lines.

**Test Cases:**
- Move line down
- Move line up
- Not move first line up (boundary condition)
- Not move last line down (boundary condition)
- Move line with children (hierarchical move)

### 9. Collapse/Expand Functionality
Tests the Shift+Click functionality for showing/hiding children.

**Test Cases:**
- Identify lines that have children
- Update display with collapse indicators (▼ for expanded, ▶ for collapsed)
- Toggle collapse state on/off
- Hide children when collapsed

### 10. Edge Cases
Tests unusual or boundary conditions to ensure robustness.

**Test Cases:**
- Handle empty content
- Handle single empty line
- Handle cursor at beginning of document
- Handle cursor at end of document
- Handle very deep nesting (5+ levels)
- Handle lines with only spaces
- Sync from display removing collapse/expand indicators
- Handle multiple consecutive indents/unindents

### 11. Cursor Position Management
Tests that cursor position is correctly maintained during operations.

**Test Cases:**
- Maintain cursor position after indent (adjusted for added spaces)
- Get current line index correctly for multiline documents

## Test Coverage

The test suite provides comprehensive coverage of:

- ✅ **Core outline operations**: indent, unindent, move up/down
- ✅ **Keyboard event handling**: Enter, Tab, Shift+Tab, Alt+Up/Down
- ✅ **Hierarchy management**: parent-child relationships, descendants
- ✅ **Cursor position tracking**: maintaining cursor during edits
- ✅ **Content conversion**: HTML ↔ plain text
- ✅ **Edge cases**: empty content, boundaries, deep nesting
- ✅ **Collapse/expand features**: show/hide children
- ✅ **Bug fixes**: Enter+Tab behavior

## Known Issues Addressed

### Enter + Tab Bug
**Issue**: When you type return and then tab, the line ABOVE the current line was being indented instead of the current line.

**Status**: Test cases created to verify correct behavior. All tests currently pass, suggesting the implementation is correct. If the bug still occurs in the live application, it may be related to:
1. Event timing/sequencing in the actual browser
2. Browser-specific selection/range behavior
3. Interaction with `updateDisplay()` that isn't captured in tests

**Recommendation**: If the bug persists in production, add integration tests that simulate actual keyboard events in a real browser environment using tools like Playwright or Cypress.

## Adding New Tests

To add a new test case:

1. Identify which category it belongs to (or create a new `describe` block)
2. Follow the pattern:
   ```javascript
   test('should describe what it tests', () => {
       // Arrange: Set up the test state
       manager.fullContent = 'Test content';
       editor.textContent = manager.fullContent;
       
       // Act: Perform the operation
       manager.someOperation();
       
       // Assert: Verify the results
       expect(manager.fullContent).toBe('Expected result');
   });
   ```
3. Run `npm test` to verify the new test works
4. Update this documentation if adding a new category

## Future Enhancements

Potential areas for additional testing:

1. **Performance tests**: Test with very large outlines (1000+ lines)
2. **Integration tests**: Test in actual browsers with real DOM
3. **Accessibility tests**: Verify keyboard navigation and screen reader support
4. **Concurrent editing**: Test behavior with rapid keystrokes
5. **Undo/Redo**: Add tests for Ctrl+Z/Ctrl+Y functionality
6. **Template operations**: Test saving and loading templates
7. **Export functionality**: Test Markdown and text export features

## Debugging Test Failures

If a test fails:

1. Read the error message - Jest provides detailed output
2. Check the test description to understand what's being tested
3. Run a single test with: `npm test -- -t "test name"`
4. Add `console.log()` statements to see intermediate values
5. Check if the DOM state is set up correctly in `beforeEach()`
6. Verify that cursor position is set correctly before operations

## Continuous Integration

These tests can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run Frontend Tests
  run: |
    npm install
    npm test
```

## Maintainer Notes

- Tests use jsdom which may not perfectly match all browser behaviors
- The `outliner.module.js` file is a wrapper to make the class testable
- Mock DOM setup happens in `beforeEach()` and cleanup in `afterEach()`
- Selection/Range API is used to simulate cursor positioning
- Some operations require `syncFromDisplay()` to be called first
