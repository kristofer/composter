/**
 * Module wrapper for outliner.js to make it testable
 * This file loads the OutlineManager class and exports it for testing
 */

const fs = require('fs');
const path = require('path');

// Read the outliner.js file
const outlinerCode = fs.readFileSync(path.join(__dirname, 'outliner.js'), 'utf8');

// Extract the OutlineManager class
let classStart = outlinerCode.indexOf('class OutlineManager');
if (classStart === -1) {
    throw new Error('Could not find OutlineManager class');
}

let braceCount = 0;
let inClass = false;
let classEnd = classStart;

for (let i = classStart; i < outlinerCode.length; i++) {
    if (outlinerCode[i] === '{') {
        braceCount++;
        inClass = true;
    } else if (outlinerCode[i] === '}') {
        braceCount--;
        if (inClass && braceCount === 0) {
            classEnd = i + 1;
            break;
        }
    }
}

const classCode = outlinerCode.substring(classStart, classEnd);

// Evaluate the class code in a way that we can export it
const moduleExports = {};
const func = new Function('module', 'exports', classCode + '\nmodule.exports = OutlineManager;');
func(moduleExports, moduleExports);

module.exports = moduleExports.exports || OutlineManager;
