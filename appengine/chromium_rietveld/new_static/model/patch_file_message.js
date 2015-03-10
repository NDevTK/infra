// Copyright (c) 2014 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

"use strict";

function PatchFileMessage(file)
{
    this.file = file || null;
    this.author = null; // User
    this.text = "";
    this.draft = false;
    this.line = 0;
    this.date = ""; // Date
    this.left = false;
    this.messageId = "";
    Object.preventExtensions(this);
}

PatchFileMessage.prototype.parseData = function(data)
{
    this.author = User.forName(data.author, data.author_email);
    this.text = (data.text || "").trim();
    this.draft = data.draft || false;
    this.line = data.lineno || 0;
    this.date = Date.utc.create(data.date);
    this.left = data.left || false;
    this.messageId = data.message_id;
};
