import showdown from 'showdown';
import marked from 'marked';
import markdownit from 'markdown-it';


import DOMPurify from 'dompurify';
import sanitizeHtml from 'sanitize-html';
import xss from 'xss';

// const shouldRenderMarkdown = (options = { project_P1, __issue_P3__, __user_P3__, issueToggle_P1.5, __commentToggle_P3__, previewToggle_P2 }) => {
const shouldRenderMarkdown = (options) => {
  if (__commentToggle_P3__) {
    return true;
  } else if (issueToggle) {
    return true;
  } else if (issue.shouldRenderMarkdown) {
    return true;
  } else if (project in allowList) {
    return true;
  } else if (__user__.preferences.defaultToMarkdown) {
    return true;
  }
  return false;
}


const showdownConverter = new showdown.Converter();
const mdit = new markdownit({ html: true, linkify: true, typographer: true });

/**
 * [description]
 * @param {[type]} content Raw content string, not HTML.
 * @param {String} options.mdLibrary enum of markdown library option.
 * @param {String} options.sanitize boolean indicating whether we sanitize markdown output.
 * @param {String} options.xssLibrary enum of xss prevention library option.
 * @return {[type]} HTML fragment string.
 */
export const renderMarkdown = (content, { mdLibrary = 'showdown', sanitize = true, xssLibrary = 'js-xss' } = {}) => {
  let converted;
  switch (mdLibrary) {
    case 'showdown':
      converted = showdownConverter.makeHtml(content);
      break;
    case 'marked':
      converted = marked(content);
      break;
    case 'markdown-it':
      converted = mdit.render(content);
      break;
    default:
      converted = showdownConverter.makeHtml(content);
      break;
  }

  if (sanitize) {
    return sanitizeMarkdown(converted, { xssLibrary });
  } else {
    return converted
  }
}

export const sanitizeMarkdown = (unsanitized, { xssLibrary = 'js-xss' } = {}) => {
  let sanitized
  switch (xssLibrary) {
    case 'js-xss':
       console.log('dedicated sanitize using js-xss')
      // return xss(html, xssOptions);
      sanitized = xss(unsanitized);
      break;
    case 'sanitize-html':
      sanitized = sanitizeHtml(unsanitized);
      break;
    case 'DOMPurify':
      sanitized = DOMPurify.sanitize(unsanitized);
      break;
    default:
      sanitized = xss(unsanitized);
  }
  return sanitized;
}


// let html
// let content = '# hello, markdown!';


// showdown
// const showdownConverter = new showdown.Converter(),
// html = showdownConverter.makeHtml(content);
// console.log(html)


// Marked
// html = marked('# Marked in Node.js\n\nRendered by **marked**.');
// html = marked(content);
// console.log(html)


// Markdown-it
// const mdit = new markdownit();
// html = mdit.render('# markdown-it rulezz!');
// html = mdit.render(content);
// console.log(html)



// DOMPurify
// import DOMPurify from 'dompurify';
// var clean = DOMPurify.sanitize(dirty);

// sanitize-html
// import sanitizeHtml from 'sanitize-html';
// const dirty = 'some really tacky HTML';
// const clean = sanitizeHtml(dirty);



// js-xss
// const xss = require("xss");
// var html = xss('<script>alert("xss");</script>');
// console.log(html);