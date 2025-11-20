/**
 * Tests for Composter Outliner
 * 
 * This test suite covers the core functionality of the OutlineManager class,
 * including keyboard shortcuts, line manipulation, and edge cases.
 */

// Load the OutlineManager class
const OutlineManager = require('./outliner.module.js');

describe('OutlineManager', () => {
    let container;
    let editor;
    let titleInput;
    let manager;

    beforeEach(() => {
        // Set up DOM elements
        document.body.innerHTML = '';
        container = document.createElement('div');
        editor = document.createElement('div');
        editor.id = 'outline-editor';
        editor.contentEditable = 'true';
        
        titleInput = document.createElement('input');
        titleInput.id = 'title';
        titleInput.type = 'text';
        
        container.appendChild(editor);
        container.appendChild(titleInput);
        document.body.appendChild(container);
        
        // Create manager instance
        manager = new OutlineManager(editor, titleInput, 0);
    });

    afterEach(() => {
        document.body.innerHTML = '';
    });

    describe('Basic Setup', () => {
        test('should initialize with empty content', () => {
            expect(manager.fullContent).toBe('');
            expect(manager.collapsedLines.size).toBe(0);
        });

        test('should initialize with provided content', () => {
            const htmlContent = '<div>Line 1</div><div style="margin-left: 30px;">Line 2</div>';
            editor.setAttribute('data-initial-content', htmlContent);
            
            const newManager = new OutlineManager(editor, titleInput, 0);
            expect(newManager.fullContent).toContain('Line 1');
            expect(newManager.fullContent).toContain('Line 2');
        });
    });

    describe('Indentation Level Calculation', () => {
        test('should calculate indent level correctly for no indentation', () => {
            expect(manager.getIndentLevel('No indent')).toBe(0);
        });

        test('should calculate indent level correctly for 2-space indent', () => {
            expect(manager.getIndentLevel('  One indent')).toBe(1);
        });

        test('should calculate indent level correctly for 4-space indent', () => {
            expect(manager.getIndentLevel('    Two indents')).toBe(2);
        });

        test('should calculate indent level correctly for 6-space indent', () => {
            expect(manager.getIndentLevel('      Three indents')).toBe(3);
        });

        test('should handle odd number of spaces', () => {
            // 5 spaces = floor(5/2) = 2 indents
            expect(manager.getIndentLevel('     Odd spaces')).toBe(2);
        });
    });

    describe('Line and Children Indices', () => {
        beforeEach(() => {
            manager.fullContent = 'Root\n  Child 1\n    Grandchild\n  Child 2\nAnother Root';
        });

        test('should find direct children correctly', () => {
            const lines = manager.fullContent.split('\n');
            const children = manager.getChildrenIndices(lines, 0);
            expect(children).toEqual([1, 3]); // Child 1 and Child 2
        });

        test('should find all descendants correctly', () => {
            const lines = manager.fullContent.split('\n');
            const descendants = manager.getAllDescendants(lines, 0);
            expect(descendants).toEqual([1, 2, 3]); // Child 1, Grandchild, Child 2
        });

        test('should return empty array when no children exist', () => {
            const lines = manager.fullContent.split('\n');
            const children = manager.getChildrenIndices(lines, 4);
            expect(children).toEqual([]);
        });
    });

    describe('Content Conversion', () => {
        test('should convert plain text to HTML', () => {
            const plainText = 'Line 1\n  Line 2\n    Line 3';
            const html = manager.plainTextToHtml(plainText);
            
            expect(html).toContain('margin-left: 0px');
            expect(html).toContain('margin-left: 30px');
            expect(html).toContain('margin-left: 60px');
            expect(html).toContain('Line 1');
            expect(html).toContain('Line 2');
            expect(html).toContain('Line 3');
        });

        test('should convert HTML to plain text', () => {
            const html = '<div>Line 1</div><div style="margin-left: 30px;">Line 2</div>';
            const plainText = manager.htmlToPlainText(html);
            
            expect(plainText).toContain('Line 1');
            expect(plainText).toContain('  Line 2');
        });

        test('should handle empty lines in conversion', () => {
            const plainText = 'Line 1\n\nLine 3';
            const html = manager.plainTextToHtml(plainText);
            expect(html).toContain('<br>');
        });
    });

    describe('Indenting Operations', () => {
        beforeEach(() => {
            manager.fullContent = 'Line 1\nLine 2\n  Child of 2';
            editor.textContent = manager.fullContent;
        });

        test('should indent a line without children', () => {
            // Set cursor on first line
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 0);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            manager.indentLineWithChildren();
            
            const lines = manager.fullContent.split('\n');
            expect(lines[0]).toBe('  Line 1');
        });

        test('should indent line with all its children', () => {
            // Set cursor on "Line 2"
            const lines = manager.fullContent.split('\n');
            // Position at start of second line
            const cursorPos = lines[0].length + 1;
            manager.setCursorPosition(cursorPos);

            manager.indentLineWithChildren();
            
            const newLines = manager.fullContent.split('\n');
            expect(newLines[1]).toBe('  Line 2');
            expect(newLines[2]).toBe('    Child of 2'); // Should also be indented
        });

        test('should unindent a line that has indentation', () => {
            manager.fullContent = '  Indented Line';
            editor.textContent = manager.fullContent;
            
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 0);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            manager.unindentLineWithChildren();
            
            expect(manager.fullContent).toBe('Indented Line');
        });

        test('should not unindent a line with no indentation', () => {
            manager.fullContent = 'No indent';
            editor.textContent = manager.fullContent;
            
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 0);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            manager.unindentLineWithChildren();
            
            expect(manager.fullContent).toBe('No indent');
        });

        test('should unindent line with all its children', () => {
            manager.fullContent = '  Line\n    Child';
            editor.textContent = manager.fullContent;
            
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 0);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            manager.unindentLineWithChildren();
            
            const lines = manager.fullContent.split('\n');
            expect(lines[0]).toBe('Line');
            expect(lines[1]).toBe('  Child'); // Should also be unindented
        });
    });

    describe('Enter Key - New Line Insertion', () => {
        test('should insert new line with same indentation at cursor position', () => {
            manager.fullContent = 'Line 1';
            editor.textContent = manager.fullContent;
            
            // Set cursor at end of line
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 6);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            manager.insertNewLineWithIndent();
            
            const lines = manager.fullContent.split('\n');
            expect(lines.length).toBe(2);
            expect(lines[0]).toBe('Line 1');
            expect(lines[1]).toBe('');
        });

        test('should insert new line with same indentation for indented line', () => {
            manager.fullContent = '  Indented Line';
            editor.textContent = manager.fullContent;
            
            // Set cursor at end
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 15);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            manager.insertNewLineWithIndent();
            
            const lines = manager.fullContent.split('\n');
            expect(lines.length).toBe(2);
            expect(lines[1]).toBe('  '); // Should have same indentation
        });

        test('should insert new line in middle of text', () => {
            manager.fullContent = 'First Second';
            editor.textContent = manager.fullContent;
            
            // Set cursor after "First "
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 6);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            manager.insertNewLineWithIndent();
            
            const lines = manager.fullContent.split('\n');
            expect(lines.length).toBe(2);
            expect(lines[0]).toBe('First ');
            expect(lines[1]).toBe('Second');
        });
    });

    describe('Enter + Tab Bug Fix', () => {
        test('after Enter, Tab should indent the new line, not the previous line', () => {
            // Start with a simple line
            manager.fullContent = 'Line 1';
            editor.textContent = manager.fullContent;
            
            // Set cursor at end of line
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 6);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            // Simulate Enter key
            manager.insertNewLineWithIndent();
            
            // After Enter, cursor should be on new line (line index 1)
            const currentLine = manager.getCurrentLineIndex();
            expect(currentLine).toBe(1);
            
            // Now simulate Tab key
            manager.indentLineWithChildren();
            
            // Check that the NEW line (line 1) is indented, not the old line (line 0)
            const lines = manager.fullContent.split('\n');
            expect(lines[0]).toBe('Line 1'); // Original line should NOT be indented
            expect(lines[1]).toBe('  '); // New line should be indented
        });

        test('after Enter on indented line, Tab should indent the new line', () => {
            manager.fullContent = '  Indented';
            editor.textContent = manager.fullContent;
            
            // Set cursor at end
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 10);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            // Press Enter
            manager.insertNewLineWithIndent();
            
            // Press Tab
            manager.indentLineWithChildren();
            
            // Check results
            const lines = manager.fullContent.split('\n');
            expect(lines[0]).toBe('  Indented'); // Original should stay the same
            expect(lines[1]).toBe('    '); // New line should have 4 spaces (2 inherited + 2 from Tab)
        });
    });

    describe('Enter Key with Collapse Indicators', () => {
        test('should insert newline correctly when collapse indicators are present', () => {
            // Set up outline with hierarchy
            manager.fullContent = 'and another line\n  some more text\n    and some more text\nthis is a second line\n  foo\nbar\nthird line\nfoo bar baz\nfoo again';
            manager.updateDisplay();
            
            // The display will have indicators for lines with children
            // "and another line" -> "▼ and another line"
            // "  some more text" -> "  ▼ some more text"
            // "this is a second line" -> "▼ this is a second line"
            const displayText = editor.textContent;
            expect(displayText).toContain('▼ and another line');
            expect(displayText).toContain('  ▼ some more text');
            expect(displayText).toContain('▼ this is a second line');
            
            // Find position at end of "foo bar baz" in the DISPLAY
            const displayLines = displayText.split('\n');
            let targetLineIndex = -1;
            for (let i = 0; i < displayLines.length; i++) {
                if (displayLines[i] === 'foo bar baz') {
                    targetLineIndex = i;
                    break;
                }
            }
            expect(targetLineIndex).toBeGreaterThanOrEqual(0);
            
            // Calculate cursor position at end of "foo bar baz" in display
            let displayCursorPos = 0;
            for (let i = 0; i < targetLineIndex; i++) {
                displayCursorPos += displayLines[i].length + 1;
            }
            displayCursorPos += displayLines[targetLineIndex].length;
            
            // Set cursor position in display
            manager.setCursorPosition(displayCursorPos);
            
            // Press Enter
            manager.insertNewLineWithIndent();
            
            // Check result
            const resultLines = manager.fullContent.split('\n');
            const bazIndex = resultLines.indexOf('foo bar baz');
            expect(bazIndex).toBeGreaterThanOrEqual(0);
            expect(resultLines[bazIndex + 1]).toBe(''); // New empty line
            expect(resultLines[bazIndex + 2]).toBe('foo again'); // Original next line should be intact
        });

        test('should insert newline correctly at end of line with children', () => {
            manager.fullContent = 'parent\n  child\nother';
            manager.updateDisplay();
            
            // Display should have: "▼ parent\n  child\nother"
            const displayText = editor.textContent;
            expect(displayText).toContain('▼ parent');
            
            // Find position at end of "▼ parent" in display
            const displayLines = displayText.split('\n');
            const parentLine = displayLines[0];
            expect(parentLine).toBe('▼ parent');
            
            // Set cursor at end of first line (after "parent")
            manager.setCursorPosition(parentLine.length);
            
            // Press Enter
            manager.insertNewLineWithIndent();
            
            // Check result - new line should be inserted after "parent", before "  child"
            const resultLines = manager.fullContent.split('\n');
            expect(resultLines[0]).toBe('parent');
            expect(resultLines[1]).toBe(''); // New empty line
            expect(resultLines[2]).toBe('  child'); // Child should still be there
            expect(resultLines[3]).toBe('other');
        });

        test('should insert newline correctly in middle of text with indicators', () => {
            manager.fullContent = 'parent line\n  child\nother text';
            manager.updateDisplay();
            
            // Display: "▼ parent line\n  child\nother text"
            const displayText = editor.textContent;
            
            // Set cursor after "parent " (in the middle of "parent line")
            // In display: "▼ parent line" -> cursor after "parent " means position 9 (▼ + space + parent + space)
            manager.setCursorPosition(9);
            
            // Press Enter
            manager.insertNewLineWithIndent();
            
            // Check result
            const resultLines = manager.fullContent.split('\n');
            expect(resultLines[0]).toBe('parent ');
            expect(resultLines[1]).toBe('line'); // Rest of text on new line
            expect(resultLines[2]).toBe('  child'); // Children preserved
            expect(resultLines[3]).toBe('other text');
        });

        test('should handle multiple levels of nesting with indicators', () => {
            manager.fullContent = 'level1\n  level2\n    level3\n      level4\nlast';
            manager.updateDisplay();
            
            // Display will have multiple indicators
            const displayText = editor.textContent;
            expect(displayText).toContain('▼ level1');
            expect(displayText).toContain('  ▼ level2');
            expect(displayText).toContain('    ▼ level3');
            
            // Set cursor at end of "last" line
            const displayLines = displayText.split('\n');
            let cursorPos = 0;
            for (let i = 0; i < displayLines.length - 1; i++) {
                cursorPos += displayLines[i].length + 1;
            }
            cursorPos += displayLines[displayLines.length - 1].length;
            
            manager.setCursorPosition(cursorPos);
            
            // Press Enter
            manager.insertNewLineWithIndent();
            
            // Check result
            const resultLines = manager.fullContent.split('\n');
            expect(resultLines.length).toBe(6); // 5 original + 1 new
            expect(resultLines[0]).toBe('level1');
            expect(resultLines[1]).toBe('  level2');
            expect(resultLines[2]).toBe('    level3');
            expect(resultLines[3]).toBe('      level4');
            expect(resultLines[4]).toBe('last');
            expect(resultLines[5]).toBe(''); // New line after "last"
        });
    });

    describe('Moving Lines', () => {
        beforeEach(() => {
            manager.fullContent = 'Line 1\nLine 2\nLine 3';
            editor.textContent = manager.fullContent;
        });

        test('should move line down', () => {
            // Set cursor on first line
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 0);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            manager.moveLineDown();
            
            const lines = manager.fullContent.split('\n');
            expect(lines[0]).toBe('Line 2');
            expect(lines[1]).toBe('Line 1');
            expect(lines[2]).toBe('Line 3');
        });

        test('should move line up', () => {
            // Set cursor on second line
            manager.setCursorPosition(7); // Position at start of "Line 2"

            manager.moveLineUp();
            
            const lines = manager.fullContent.split('\n');
            expect(lines[0]).toBe('Line 2');
            expect(lines[1]).toBe('Line 1');
            expect(lines[2]).toBe('Line 3');
        });

        test('should not move first line up', () => {
            const originalContent = manager.fullContent;
            
            // Set cursor on first line
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 0);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            manager.moveLineUp();
            
            expect(manager.fullContent).toBe(originalContent);
        });

        test('should not move last line down', () => {
            const originalContent = manager.fullContent;
            
            // Set cursor on last line
            manager.setCursorPosition(14); // Position at start of "Line 3"

            manager.moveLineDown();
            
            expect(manager.fullContent).toBe(originalContent);
        });

        test('should move line with children', () => {
            manager.fullContent = 'Line 1\nParent\n  Child\nLine 4';
            editor.textContent = manager.fullContent;
            
            // Move "Parent" and its child down
            manager.setCursorPosition(7); // Position at "Parent"
            manager.moveLineDown();
            
            const lines = manager.fullContent.split('\n');
            expect(lines[0]).toBe('Line 1');
            expect(lines[1]).toBe('Line 4');
            expect(lines[2]).toBe('Parent');
            expect(lines[3]).toBe('  Child');
        });
    });

    describe('Collapse/Expand Functionality', () => {
        beforeEach(() => {
            manager.fullContent = 'Parent\n  Child 1\n  Child 2\nOther';
        });

        test('should identify lines with children', () => {
            const lines = manager.fullContent.split('\n');
            const children = manager.getChildrenIndices(lines, 0);
            expect(children.length).toBe(2);
        });

        test('should update display with collapse indicators', () => {
            manager.updateDisplay();
            
            const displayText = editor.textContent;
            expect(displayText).toContain('▼'); // Expanded indicator
        });

        test('should toggle collapse state', () => {
            manager.updateDisplay();
            editor.textContent = manager.fullContent;
            
            // Set cursor on parent line
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 0);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            manager.toggleChildren(new MouseEvent('click'));
            
            expect(manager.collapsedLines.has(0)).toBe(true);
            
            manager.toggleChildren(new MouseEvent('click'));
            
            expect(manager.collapsedLines.has(0)).toBe(false);
        });

        test('should hide children when collapsed', () => {
            manager.collapsedLines.add(0);
            manager.updateDisplay();
            
            const displayText = editor.textContent;
            expect(displayText).toContain('▶'); // Collapsed indicator
            expect(displayText).not.toContain('Child 1');
            expect(displayText).not.toContain('Child 2');
        });
    });

    describe('Edge Cases', () => {
        test('should handle empty content', () => {
            manager.fullContent = '';
            expect(() => manager.updateDisplay()).not.toThrow();
        });

        test('should handle single empty line', () => {
            manager.fullContent = '\n';
            const lines = manager.fullContent.split('\n');
            expect(lines.length).toBe(2);
        });

        test('should handle cursor at beginning of document', () => {
            manager.fullContent = 'Test';
            editor.textContent = manager.fullContent;
            
            manager.setCursorPosition(0);
            expect(manager.getCurrentLineIndex()).toBe(0);
        });

        test('should handle cursor at end of document', () => {
            manager.fullContent = 'Line 1\nLine 2';
            editor.textContent = manager.fullContent;
            
            const lastPos = manager.fullContent.length;
            manager.setCursorPosition(lastPos);
            expect(manager.getCurrentLineIndex()).toBe(1);
        });

        test('should handle very deep nesting', () => {
            manager.fullContent = 'L0\n  L1\n    L2\n      L3\n        L4\n          L5';
            const lines = manager.fullContent.split('\n');
            
            expect(manager.getIndentLevel(lines[5])).toBe(5);
            const descendants = manager.getAllDescendants(lines, 0);
            expect(descendants.length).toBe(5);
        });

        test('should handle lines with only spaces', () => {
            manager.fullContent = 'Line 1\n    \nLine 3';
            const lines = manager.fullContent.split('\n');
            expect(manager.getIndentLevel(lines[1])).toBe(2);
        });

        test('should sync from display removing indicators', () => {
            editor.textContent = '▼ Parent\n  Child';
            manager.syncFromDisplay();
            
            expect(manager.fullContent).not.toContain('▼');
            expect(manager.fullContent).toContain('Parent');
        });

        test('should handle multiple consecutive indents/unindents', () => {
            manager.fullContent = 'Line';
            editor.textContent = manager.fullContent;
            
            const range = document.createRange();
            const textNode = editor.firstChild;
            range.setStart(textNode, 0);
            range.collapse(true);
            window.getSelection().removeAllRanges();
            window.getSelection().addRange(range);

            // Indent 3 times
            manager.indentLineWithChildren();
            manager.indentLineWithChildren();
            manager.indentLineWithChildren();
            
            expect(manager.fullContent).toBe('      Line');
            
            // Unindent 3 times
            manager.unindentLineWithChildren();
            manager.unindentLineWithChildren();
            manager.unindentLineWithChildren();
            
            expect(manager.fullContent).toBe('Line');
        });
    });

    describe('Cursor Position Management', () => {
        test('should maintain cursor position after indent', () => {
            manager.fullContent = 'Test Line';
            editor.textContent = manager.fullContent;
            
            // Set cursor in middle
            manager.setCursorPosition(5);
            const initialPos = manager.getCursorPosition();
            
            manager.indentLineWithChildren();
            
            const newPos = manager.getCursorPosition();
            expect(newPos).toBe(initialPos + 2); // Moved by 2 spaces
        });

        test('should get current line index correctly for multiline', () => {
            manager.fullContent = 'Line 1\nLine 2\nLine 3';
            editor.textContent = manager.fullContent;
            
            // Position at start of second line
            manager.setCursorPosition(7);
            expect(manager.getCurrentLineIndex()).toBe(1);
            
            // Position at start of third line  
            manager.setCursorPosition(14);
            expect(manager.getCurrentLineIndex()).toBe(2);
        });
    });
});
