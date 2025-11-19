/**
 * Composter Outliner - A keyboard-driven hierarchical outliner
 * Optimized for program/task decomposition
 */

class OutlineManager {
    constructor(editorElement, titleElement, outlineId) {
        this.editor = editorElement;
        this.titleInput = titleElement;
        this.outlineId = outlineId;
        this.collapsedLines = new Set();
        this.fullContent = '';
        
        this.initializeEditor();
        this.attachEventListeners();
    }

    initializeEditor() {
        // Load initial content if exists
        const initialContent = this.editor.getAttribute('data-initial-content') || '';
        if (initialContent) {
            this.fullContent = this.htmlToPlainText(initialContent);
            this.updateDisplay();
        }
    }

    attachEventListeners() {
        // Click handlers for collapse/expand
        this.editor.addEventListener('click', (e) => {
            if (e.shiftKey) {
                e.preventDefault();
                if (e.ctrlKey || e.metaKey) {
                    this.toggleAllCollapse();
                } else {
                    this.toggleChildren(e);
                }
            }
        });

        // Keyboard shortcuts
        this.editor.addEventListener('keydown', (e) => this.handleKeyDown(e));
        
        // Global keyboard shortcuts (for help overlay)
        document.addEventListener('keydown', (e) => {
            if (e.key === '?' && !e.ctrlKey && !e.metaKey && !e.altKey) {
                // Only show help if not typing in an input/textarea
                if (e.target.contentEditable !== 'true' && 
                    e.target.tagName !== 'INPUT' && 
                    e.target.tagName !== 'TEXTAREA') {
                    return;
                }
                e.preventDefault();
                this.toggleHelpOverlay();
            }
            if (e.key === 'Escape') {
                this.hideHelpOverlay();
            }
        });
    }

    handleKeyDown(e) {
        // Tab for indent
        if (e.key === 'Tab') {
            e.preventDefault();
            if (e.shiftKey) {
                this.unindentLineWithChildren();
            } else {
                this.indentLineWithChildren();
            }
        }
        // Alt+Up: Move item up (with children)
        else if (e.altKey && e.key === 'ArrowUp') {
            e.preventDefault();
            this.moveLineUp();
        }
        // Alt+Down: Move item down (with children)
        else if (e.altKey && e.key === 'ArrowDown') {
            e.preventDefault();
            this.moveLineDown();
        }
        // Ctrl/Cmd + S for save
        else if ((e.ctrlKey || e.metaKey) && e.key === 's') {
            e.preventDefault();
            this.save(false);
        }
        // Enter key - new line with same indentation
        else if (e.key === 'Enter') {
            e.preventDefault();
            this.insertNewLineWithIndent();
        }
    }

    // ========== Utility Methods ==========

    /**
     * Get the absolute cursor position in the editor
     */
    getCursorPosition() {
        const selection = window.getSelection();
        if (!selection.rangeCount) return 0;
        
        const range = selection.getRangeAt(0);
        const preCaretRange = range.cloneRange();
        preCaretRange.selectNodeContents(this.editor);
        preCaretRange.setEnd(range.endContainer, range.endOffset);
        return preCaretRange.toString().length;
    }

    /**
     * Set the cursor to a specific position
     */
    setCursorPosition(pos) {
        const selection = window.getSelection();
        const range = document.createRange();
        
        let currentPos = 0;
        let found = false;
        
        const traverse = (node) => {
            if (found) return;
            
            if (node.nodeType === Node.TEXT_NODE) {
                const nodeLength = node.textContent.length;
                if (currentPos + nodeLength >= pos) {
                    range.setStart(node, pos - currentPos);
                    range.collapse(true);
                    found = true;
                    return;
                }
                currentPos += nodeLength;
            } else {
                for (let child of node.childNodes) {
                    traverse(child);
                    if (found) return;
                }
            }
        };
        
        traverse(this.editor);
        
        if (found) {
            selection.removeAllRanges();
            selection.addRange(range);
        }
    }

    /**
     * Get the current line index based on cursor position
     */
    getCurrentLineIndex() {
        const text = this.editor.textContent;
        const lines = text.split('\n');
        const cursorPos = this.getCursorPosition();
        
        let charCount = 0;
        for (let i = 0; i < lines.length; i++) {
            charCount += lines[i].length;
            if (i < lines.length - 1) charCount++; // Add newline
            if (cursorPos < charCount || (cursorPos === charCount && i === lines.length - 1)) {
                return i;
            }
        }
        return 0;
    }

    /**
     * Get indentation level of a line (number of 2-space indents)
     */
    getIndentLevel(lineText) {
        let spaces = 0;
        for (let i = 0; i < lineText.length; i++) {
            if (lineText[i] === ' ') {
                spaces++;
            } else {
                break;
            }
        }
        return Math.floor(spaces / 2);
    }

    /**
     * Get all direct children indices for a parent line
     */
    getChildrenIndices(lines, parentIndex) {
        const parentIndent = this.getIndentLevel(lines[parentIndex]);
        const children = [];
        
        for (let i = parentIndex + 1; i < lines.length; i++) {
            const indent = this.getIndentLevel(lines[i]);
            if (indent <= parentIndent) break;
            if (indent === parentIndent + 1) {
                children.push(i);
            }
        }
        
        return children;
    }

    /**
     * Get all descendant indices (children, grandchildren, etc.)
     */
    getAllDescendants(lines, parentIndex) {
        const parentIndent = this.getIndentLevel(lines[parentIndex]);
        const descendants = [];
        
        for (let i = parentIndex + 1; i < lines.length; i++) {
            const indent = this.getIndentLevel(lines[i]);
            if (indent <= parentIndent) break;
            descendants.push(i);
        }
        
        return descendants;
    }

    /**
     * Sync fullContent from editor display (remove indicators)
     */
    syncFromDisplay() {
        const currentText = this.editor.textContent;
        const lines = currentText.split('\n');
        
        // Remove collapse/expand indicators from lines
        const cleanedLines = lines.map(line => {
            let spaces = 0;
            for (let i = 0; i < line.length; i++) {
                if (line[i] === ' ') {
                    spaces++;
                } else {
                    break;
                }
            }
            const indent = ' '.repeat(spaces);
            let rest = line.substring(spaces);
            
            // Remove indicators
            while (rest.startsWith('▶ ') || rest.startsWith('▼ ')) {
                rest = rest.substring(2);
            }
            
            return indent + rest;
        });
        
        this.fullContent = cleanedLines.join('\n');
    }

    /**
     * Update the editor display with collapse indicators
     */
    updateDisplay() {
        const lines = this.fullContent.split('\n');
        
        // Build set of hidden lines
        const hiddenLines = new Set();
        this.collapsedLines.forEach(parentIndex => {
            const descendants = this.getAllDescendants(lines, parentIndex);
            descendants.forEach(index => hiddenLines.add(index));
        });
        
        // Render visible lines with indicators
        const visibleLines = lines.map((line, index) => {
            if (hiddenLines.has(index)) return null;
            
            const children = this.getChildrenIndices(lines, index);
            if (children.length > 0) {
                const isCollapsed = this.collapsedLines.has(index);
                const indicator = isCollapsed ? '▶ ' : '▼ ';
                
                let spaces = 0;
                for (let i = 0; i < line.length; i++) {
                    if (line[i] === ' ') {
                        spaces++;
                    } else {
                        break;
                    }
                }
                const indent = ' '.repeat(spaces);
                const content = line.substring(spaces);
                return indent + indicator + content;
            }
            
            return line;
        }).filter(line => line !== null);
        
        this.editor.textContent = visibleLines.join('\n');
    }

    /**
     * Convert HTML content to plain text with indentation
     */
    htmlToPlainText(html) {
        const temp = document.createElement('div');
        temp.innerHTML = html;
        
        const lines = [];
        const divs = temp.querySelectorAll('div');
        
        if (divs.length > 0) {
            divs.forEach(div => {
                const marginLeft = parseInt(div.style.marginLeft || '0');
                const indentLevel = Math.floor(marginLeft / 30);
                const indent = '  '.repeat(indentLevel);
                const text = div.textContent || '';
                lines.push(indent + text);
            });
            return lines.join('\n');
        }
        
        return temp.textContent || '';
    }

    /**
     * Convert plain text to HTML for storage
     */
    plainTextToHtml(plainText) {
        const lines = plainText.split('\n');
        return lines.map(line => {
            let spaces = 0;
            for (let i = 0; i < line.length; i++) {
                if (line[i] === ' ') {
                    spaces++;
                } else {
                    break;
                }
            }
            const indentLevel = Math.floor(spaces / 2);
            const marginLeft = indentLevel * 30;
            const text = line.trim() || '<br>';
            return `<div style="margin-left: ${marginLeft}px;">${text}</div>`;
        }).join('');
    }

    /**
     * Show a temporary message
     */
    showMessage(text, type = 'success') {
        const msg = document.createElement('div');
        msg.textContent = text;
        msg.className = `toast-notification ${type}`;
        document.body.appendChild(msg);
        
        setTimeout(() => {
            msg.style.animation = 'slideOutRight 0.3s ease-out';
            setTimeout(() => msg.remove(), 300);
        }, 1700);
    }

    /**
     * Toggle keyboard shortcut help overlay
     */
    toggleHelpOverlay() {
        const existing = document.getElementById('shortcut-help');
        if (existing) {
            this.hideHelpOverlay();
        } else {
            this.showHelpOverlay();
        }
    }

    /**
     * Show keyboard shortcut help overlay
     */
    showHelpOverlay() {
        const overlay = document.createElement('div');
        overlay.id = 'shortcut-help';
        overlay.className = 'shortcut-help-overlay';
        
        overlay.innerHTML = `
            <div class="shortcut-help-content">
                <h2>⌨️ Keyboard Shortcuts</h2>
                
                <h3>Outlining</h3>
                <ul class="shortcut-list">
                    <li class="shortcut-item">
                        <span>Indent line (with children)</span>
                        <div class="shortcut-keys"><kbd>Tab</kbd></div>
                    </li>
                    <li class="shortcut-item">
                        <span>Unindent line (with children)</span>
                        <div class="shortcut-keys"><kbd>Shift</kbd> + <kbd>Tab</kbd></div>
                    </li>
                    <li class="shortcut-item">
                        <span>Move item up (with children)</span>
                        <div class="shortcut-keys"><kbd>Alt</kbd> + <kbd>↑</kbd></div>
                    </li>
                    <li class="shortcut-item">
                        <span>Move item down (with children)</span>
                        <div class="shortcut-keys"><kbd>Alt</kbd> + <kbd>↓</kbd></div>
                    </li>
                    <li class="shortcut-item">
                        <span>New line (same indent)</span>
                        <div class="shortcut-keys"><kbd>Enter</kbd></div>
                    </li>
                </ul>
                
                <h3>View Control</h3>
                <ul class="shortcut-list">
                    <li class="shortcut-item">
                        <span>Collapse/expand children</span>
                        <div class="shortcut-keys"><kbd>Shift</kbd> + <kbd>Click</kbd></div>
                    </li>
                    <li class="shortcut-item">
                        <span>Collapse/expand all</span>
                        <div class="shortcut-keys"><kbd>Ctrl</kbd> + <kbd>Shift</kbd> + <kbd>Click</kbd></div>
                    </li>
                </ul>
                
                <h3>Document</h3>
                <ul class="shortcut-list">
                    <li class="shortcut-item">
                        <span>Save outline</span>
                        <div class="shortcut-keys"><kbd>Ctrl</kbd> + <kbd>S</kbd></div>
                    </li>
                    <li class="shortcut-item">
                        <span>Undo</span>
                        <div class="shortcut-keys"><kbd>Ctrl</kbd> + <kbd>Z</kbd></div>
                    </li>
                    <li class="shortcut-item">
                        <span>Redo</span>
                        <div class="shortcut-keys"><kbd>Ctrl</kbd> + <kbd>Y</kbd></div>
                    </li>
                </ul>
                
                <h3>Help</h3>
                <ul class="shortcut-list">
                    <li class="shortcut-item">
                        <span>Show/hide this help</span>
                        <div class="shortcut-keys"><kbd>?</kbd></div>
                    </li>
                    <li class="shortcut-item">
                        <span>Close this help</span>
                        <div class="shortcut-keys"><kbd>Esc</kbd></div>
                    </li>
                </ul>
                
                <button class="close-help" onclick="document.getElementById('shortcut-help').remove()">
                    Close (or press Esc)
                </button>
            </div>
        `;
        
        document.body.appendChild(overlay);
        
        // Close on overlay click
        overlay.addEventListener('click', (e) => {
            if (e.target === overlay) {
                this.hideHelpOverlay();
            }
        });
    }

    /**
     * Hide keyboard shortcut help overlay
     */
    hideHelpOverlay() {
        const overlay = document.getElementById('shortcut-help');
        if (overlay) {
            overlay.remove();
        }
    }

    // ========== Outline Manipulation Methods ==========

    /**
     * Indent current line and all its children
     */
    indentLineWithChildren() {
        this.syncFromDisplay();
        
        const lines = this.fullContent.split('\n');
        const cursorPos = this.getCursorPosition();
        const currentLineIndex = this.getCurrentLineIndex();
        const descendants = this.getAllDescendants(lines, currentLineIndex);
        
        // Indent current line and all descendants
        [currentLineIndex, ...descendants].forEach(index => {
            lines[index] = '  ' + lines[index];
        });
        
        this.fullContent = lines.join('\n');
        this.updateDisplay();
        this.setCursorPosition(cursorPos + 2);
        this.showMessage('Indented', 'success');
    }

    /**
     * Unindent current line and all its children
     */
    unindentLineWithChildren() {
        this.syncFromDisplay();
        
        const lines = this.fullContent.split('\n');
        const cursorPos = this.getCursorPosition();
        const currentLineIndex = this.getCurrentLineIndex();
        const currentLine = lines[currentLineIndex];
        
        // Check if current line has indent to remove
        let spacesToRemove = 0;
        for (let i = 0; i < Math.min(2, currentLine.length); i++) {
            if (currentLine[i] === ' ') {
                spacesToRemove++;
            } else {
                break;
            }
        }
        
        if (spacesToRemove === 0) return;
        
        const descendants = this.getAllDescendants(lines, currentLineIndex);
        
        // Unindent current line and all descendants
        [currentLineIndex, ...descendants].forEach(index => {
            let removed = 0;
            for (let i = 0; i < Math.min(2, lines[index].length); i++) {
                if (lines[index][i] === ' ') {
                    removed++;
                } else {
                    break;
                }
            }
            if (removed > 0) {
                lines[index] = lines[index].substring(removed);
            }
        });
        
        this.fullContent = lines.join('\n');
        this.updateDisplay();
        this.setCursorPosition(Math.max(0, cursorPos - spacesToRemove));
        this.showMessage('Unindented', 'success');
    }

    /**
     * Move current line (and children) up
     */
    moveLineUp() {
        this.syncFromDisplay();
        
        const lines = this.fullContent.split('\n');
        const cursorPos = this.getCursorPosition();
        const currentLineIndex = this.getCurrentLineIndex();
        
        if (currentLineIndex === 0) return;
        
        // Calculate cursor offset within the current line before moving
        let charCount = 0;
        for (let i = 0; i < currentLineIndex; i++) {
            charCount += lines[i].length + 1;
        }
        const cursorOffsetInLine = cursorPos - charCount;
        
        const descendants = this.getAllDescendants(lines, currentLineIndex);
        const linesToMove = [currentLineIndex, ...descendants];
        
        // Find target position
        let targetIndex = currentLineIndex - 1;
        const currentIndent = this.getIndentLevel(lines[currentLineIndex]);
        const prevIndent = this.getIndentLevel(lines[targetIndex]);
        
        // Can't move if it would break the tree structure
        if (prevIndent > currentIndent) return;
        
        // If previous line has children, skip over them
        const prevDescendants = this.getAllDescendants(lines, targetIndex);
        if (prevDescendants.length > 0) {
            targetIndex = targetIndex - prevDescendants.length - 1;
            if (targetIndex < 0) return;
        }
        
        // Extract lines to move
        const movedLines = linesToMove.map(i => lines[i]);
        
        // Remove from current position
        const newLines = lines.filter((_, i) => !linesToMove.includes(i));
        
        // Calculate new position after removal
        let insertAt = targetIndex;
        for (let i = 0; i < linesToMove.length; i++) {
            if (linesToMove[i] < targetIndex) {
                insertAt--;
            }
        }
        
        // Insert at new position
        newLines.splice(insertAt, 0, ...movedLines);
        
        this.fullContent = newLines.join('\n');
        this.updateDisplay();
        
        // Restore cursor position
        let newCursorPos = 0;
        for (let i = 0; i < insertAt; i++) {
            newCursorPos += newLines[i].length + 1;
        }
        newCursorPos += cursorOffsetInLine;
        
        this.setCursorPosition(newCursorPos);
        this.showMessage('Moved up', 'success');
    }

    /**
     * Move current line (and children) down
     */
    moveLineDown() {
        this.syncFromDisplay();
        
        const lines = this.fullContent.split('\n');
        const cursorPos = this.getCursorPosition();
        const currentLineIndex = this.getCurrentLineIndex();
        
        // Calculate cursor offset
        let charCount = 0;
        for (let i = 0; i < currentLineIndex; i++) {
            charCount += lines[i].length + 1;
        }
        const cursorOffsetInLine = cursorPos - charCount;
        
        const descendants = this.getAllDescendants(lines, currentLineIndex);
        const linesToMove = [currentLineIndex, ...descendants];
        const lastIndexToMove = Math.max(...linesToMove);
        
        // Can't move if at the end
        if (lastIndexToMove >= lines.length - 1) return;
        
        const nextIndex = lastIndexToMove + 1;
        const currentIndent = this.getIndentLevel(lines[currentLineIndex]);
        const nextIndent = this.getIndentLevel(lines[nextIndex]);
        
        // Can't move if it would break the tree structure
        if (nextIndent > currentIndent) return;
        
        // Find where to insert (after next item and its descendants)
        const nextDescendants = this.getAllDescendants(lines, nextIndex);
        let insertAfter = nextIndex;
        if (nextDescendants.length > 0) {
            insertAfter = Math.max(...nextDescendants);
        }
        
        // Extract lines to move
        const movedLines = linesToMove.map(i => lines[i]);
        
        // Remove from current position
        const newLines = lines.filter((_, i) => !linesToMove.includes(i));
        
        // Calculate new position after removal
        let insertAt = insertAfter + 1;
        for (let i = 0; i < linesToMove.length; i++) {
            if (linesToMove[i] <= insertAfter) {
                insertAt--;
            }
        }
        
        // Insert at new position
        newLines.splice(insertAt, 0, ...movedLines);
        
        this.fullContent = newLines.join('\n');
        this.updateDisplay();
        
        // Restore cursor position
        let newCursorPos = 0;
        for (let i = 0; i < insertAt; i++) {
            newCursorPos += newLines[i].length + 1;
        }
        newCursorPos += cursorOffsetInLine;
        
        this.setCursorPosition(newCursorPos);
        this.showMessage('Moved down', 'success');
    }

    /**
     * Insert a new line with same indentation as current line
     */
    insertNewLineWithIndent() {
        this.syncFromDisplay();
        
        const cursorPos = this.getCursorPosition();
        const lines = this.fullContent.split('\n');
        const currentLineIndex = this.getCurrentLineIndex();
        const lineText = lines[currentLineIndex];
        
        // Count leading spaces in current line
        let spaces = 0;
        for (let i = 0; i < lineText.length; i++) {
            if (lineText[i] === ' ') {
                spaces++;
            } else {
                break;
            }
        }
        
        // Insert newline with same indentation
        const indent = ' '.repeat(spaces);
        const newText = this.fullContent.substring(0, cursorPos) + 
                       '\n' + indent + 
                       this.fullContent.substring(cursorPos);
        
        this.fullContent = newText;
        this.updateDisplay();
        this.setCursorPosition(cursorPos + 1 + spaces);
    }

    /**
     * Toggle children visibility for clicked line
     */
    toggleChildren(e) {
        const cursorPos = this.getCursorPosition();
        this.setCursorPosition(cursorPos);
        
        const lines = this.fullContent.split('\n');
        const currentLineIndex = this.getCurrentLineIndex();
        const children = this.getChildrenIndices(lines, currentLineIndex);
        
        if (children.length === 0) return;
        
        if (this.collapsedLines.has(currentLineIndex)) {
            this.collapsedLines.delete(currentLineIndex);
        } else {
            this.collapsedLines.add(currentLineIndex);
        }
        
        this.updateDisplay();
        this.setCursorPosition(cursorPos);
    }

    /**
     * Toggle all lines with children (collapse all or expand all)
     */
    toggleAllCollapse() {
        const lines = this.fullContent.split('\n');
        
        if (this.collapsedLines.size > 0) {
            // Expand all
            this.collapsedLines.clear();
        } else {
            // Collapse all: find all lines with children
            for (let i = 0; i < lines.length; i++) {
                const children = this.getChildrenIndices(lines, i);
                if (children.length > 0) {
                    this.collapsedLines.add(i);
                }
            }
        }
        
        const cursorPos = this.getCursorPosition();
        this.updateDisplay();
        this.setCursorPosition(cursorPos);
    }

    // ========== Save/Load Methods ==========

    /**
     * Export outline to Markdown format
     */
    exportToMarkdown() {
        this.syncFromDisplay();
        const lines = this.fullContent.split('\n');
        
        // Convert to markdown with proper formatting
        const markdown = lines.map(line => {
            const indentLevel = this.getIndentLevel(line);
            const content = line.trim();
            
            if (!content) return '';
            
            // Use markdown list syntax
            const indent = '  '.repeat(indentLevel);
            return `${indent}- ${content}`;
        }).join('\n');
        
        // Add title if exists
        const title = this.titleInput.value.trim();
        const fullMarkdown = title ? `# ${title}\n\n${markdown}` : markdown;
        
        // Create download
        this.downloadFile(fullMarkdown, `${title || 'outline'}.md`, 'text/markdown');
        this.showMessage('Exported to Markdown!', 'success');
    }

    /**
     * Export outline to plain text format
     */
    exportToText() {
        this.syncFromDisplay();
        
        // Add title if exists
        const title = this.titleInput.value.trim();
        const fullText = title ? `${title}\n${'='.repeat(title.length)}\n\n${this.fullContent}` : this.fullContent;
        
        // Create download
        this.downloadFile(fullText, `${title || 'outline'}.txt`, 'text/plain');
        this.showMessage('Exported to text!', 'success');
    }

    /**
     * Helper to download a file
     */
    downloadFile(content, filename, mimeType) {
        const blob = new Blob([content], { type: mimeType });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    /**
     * Save the outline to the server
     */
    save(shouldClose = false) {
        const title = this.titleInput.value.trim();
        if (!title) {
            alert('Please enter a title');
            this.titleInput.focus();
            return;
        }
        
        this.syncFromDisplay();
        const htmlContent = this.plainTextToHtml(this.fullContent);
        
        fetch('/api/outline/save', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                id: this.outlineId,
                title: title,
                content: htmlContent
            })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                if (shouldClose) {
                    window.location.href = '/';
                } else {
                    this.showMessage('Saved!', 'success');
                }
            } else {
                alert('Error saving outline');
            }
        })
        .catch(error => {
            console.error('Save error:', error);
            alert('Error saving outline');
        });
    }

    /**
     * Save outline as a template
     */
    saveAsTemplate() {
        this.syncFromDisplay();
        const htmlContent = this.plainTextToHtml(this.fullContent);
        
        // Show modal to get template details
        const modal = document.createElement('div');
        modal.style.cssText = 'position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 9999;';
        
        modal.innerHTML = `
            <div style="background: white; padding: 30px; border-radius: 8px; max-width: 500px; width: 90%;">
                <h3 style="margin-top: 0;">Save as Template</h3>
                <div style="margin-bottom: 15px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 500;">Template Name:</label>
                    <input type="text" id="template-name" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;" placeholder="e.g., My Custom Pattern">
                </div>
                <div style="margin-bottom: 15px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 500;">Description:</label>
                    <textarea id="template-description" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; min-height: 80px;" placeholder="Brief description of this template"></textarea>
                </div>
                <div style="margin-bottom: 20px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 500;">Category:</label>
                    <select id="template-category" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
                        <option value="General">General</option>
                        <option value="MVC">MVC</option>
                        <option value="API">API</option>
                        <option value="Microservice">Microservice</option>
                        <option value="DataPipeline">Data Pipeline</option>
                        <option value="Feature">Feature</option>
                        <option value="BugFix">Bug Fix</option>
                    </select>
                </div>
                <div style="display: flex; gap: 10px; justify-content: flex-end;">
                    <button id="cancel-template" style="padding: 8px 16px; border: 1px solid #ddd; background: white; border-radius: 4px; cursor: pointer;">Cancel</button>
                    <button id="submit-template" style="padding: 8px 16px; border: none; background: #3498db; color: white; border-radius: 4px; cursor: pointer;">Save Template</button>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        document.getElementById('template-name').focus();
        
        document.getElementById('cancel-template').onclick = () => {
            document.body.removeChild(modal);
        };
        
        document.getElementById('submit-template').onclick = () => {
            const name = document.getElementById('template-name').value.trim();
            const description = document.getElementById('template-description').value.trim();
            const category = document.getElementById('template-category').value;
            
            if (!name) {
                alert('Please enter a template name');
                return;
            }
            
            if (!description) {
                alert('Please enter a description');
                return;
            }
            
            fetch('/api/template/create', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    name: name,
                    description: description,
                    content: htmlContent,
                    category: category
                })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    document.body.removeChild(modal);
                    this.showMessage('Template saved successfully!', 'success');
                } else {
                    alert('Error saving template');
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('Error saving template');
            });
        };
    }
}

// Initialize when DOM is ready
function initializeOutliner() {
    const editor = document.getElementById('outline-editor');
    const titleInput = document.getElementById('title');
    const outlineId = window.outlineId || 0;
    
    if (editor && titleInput) {
        const manager = new OutlineManager(editor, titleInput, outlineId);
        
        // Expose globally for button clicks and help
        window.outlinerManager = manager;
        window.saveOutline = (shouldClose) => manager.save(shouldClose);
        window.saveAsTemplate = () => manager.saveAsTemplate();
        
        // Focus the editor on load if title is filled
        if (titleInput.value) {
            setTimeout(() => editor.focus(), 100);
        }
    }
}

// Auto-initialize if DOM is already ready, otherwise wait
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeOutliner);
} else {
    initializeOutliner();
}
