// Copyright (c) 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

"use strict";

// Stateful highlighter that can be used to syntax highlight a file by calling
// parseText() repeatedly on chunks of the file (usually per line).
function SyntaxHighlighter(initialLanguage, containsEmbeddedLanguages)
{
    // State of the highlighter from hljs to make multi line comments work.
    this.syntaxState = null;
    this.initialLanguage = initialLanguage; // string.
    this.language = initialLanguage; // string.
    this.containsEmbeddedLanguages = containsEmbeddedLanguages; // boolean.
    Object.preventExtensions(this);
}

// Resets the highlighter to the intial state, for example if you're inside a
// multi line comment this will reset it back to regular statement mode.
SyntaxHighlighter.prototype.resetSyntaxState = function()
{
    this.syntaxState = null;
};

// Parse a string into syntax highlighted html. If the string cannot be
// highlighted, or no highlighting is needed, it will return null.
//
// Parsing is stateful, so parsing "/*" in C++ will mean all future text parsed
// will be highlighted like a comment until the string "*/" is encountered.
SyntaxHighlighter.prototype.parseText = function(text)
{
    if (!this.language)
        return null;

    if (this.shouldResetEmbeddedLanguage(text)) {
        this.syntaxState = null;
        this.language = this.initialLanguage;
    }

    var code = this.parseTextInternal(text);
    if (code)
        this.syntaxState = code.top;

    var language = this.selectEmbeddedLanguage(text);
    if (language) {
        this.syntaxState = null;
        this.language = language;
    }

    if (code)
        return code.value;
    return null;
};

SyntaxHighlighter.prototype.parseTextInternal = function(text)
{
    // Keep this in a separate function since v8 will de-optimize functions
    // with try/catch.
    try {
        return hljs.highlight(this.language, text, true, this.syntaxState);
    } catch (e) {
        console.log("Syntax highlighter error", e);
    }
    return null;
};

SyntaxHighlighter.prototype.selectEmbeddedLanguage = function(text)
{
    if (!this.containsEmbeddedLanguages)
        return "";
    if (text.startsWith("<script") && !text.contains("<\/script"))
        return "javascript";
    if (text.startsWith("<style") && !text.contains("<\/style"))
        return "css";
    return "";
};

SyntaxHighlighter.prototype.shouldResetEmbeddedLanguage = function(text)
{
    if (!this.containsEmbeddedLanguages)
        return false;
    if (this.language == "javascript" && text.startsWith("<\/script"))
        return true;
    if (this.language == "css" && text.startsWith("<\/style"))
        return true;
    return false;
};
