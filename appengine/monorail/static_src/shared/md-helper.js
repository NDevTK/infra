import marked from 'marked';
import DOMPurify from 'dompurify';

/** @type {Set} Projects that defaults content as Markdown content. */
const DEFAULT_MD_PROJECTS = new Set();

/** @type {Set} Authors whose comments will not be rendered as Markdown. */
const BLOCKLIST = new Set();

/**
 * Determines whether content should be rendered as Markdown.
 * @param {string} options.project Project this content belongs to.
 * @param {number} options.author User who authored this content.
 * @param {boolean} options.override Per-issue override to force Markdown.
 * @return {boolean} Whether this content should be rendered as Markdown.
 */
export const shouldRenderMarkdown = ({
  project, author, override = false,
} = {}) => {
  if (author in BLOCKLIST) {
    return false;
  } else if (override) {
    return true;
  } else if (project in DEFAULT_MD_PROJECTS) {
    return true;
  }
  return false;
};

/**
 * Replaces bold HTML tags in comment with Markdown equivalent.
 * @param {string} raw Comment string as stored in database.
 * @return {string} Comment string after b tags are placed by Markdown bolding.
 */
const replaceBoldTag = (raw) => {
  return raw.replace(/<b>/g, '**').replace(/<\/b>/g, '**');
};

/** @const {Object} Options for DOMPurify sanitizer */
const SANITIZE_OPTIONS = Object.freeze({
  RETURN_TRUSTED_TYPE: true,
  FORBID_TAGS: ['style'],
  FORBID_ATTR: ['style', 'autoplay'],
});

// TODO: set other options for preprocessor and stuff.
marked.use({headerIds: false});

/**
 * Renders Markdown content into HTML.
 * @param {string} raw Content to be intepretted as Markdown.
 * @return {TrustedHTML} Rendered content in HTML format.
 */
export const renderMarkdown = (raw) => {
  // TODO: May have to also have inputs: commentReferences, projectName,
  // and revisionUrlFormat to use in conjunction with Marked's lexer for
  // autolinking.
  const preprocessed = replaceBoldTag(raw);
  // TODO: Escape source HTML
  // TODO: Use autolink
  const converted = marked(preprocessed);
  const sanitized = DOMPurify.sanitize(converted, SANITIZE_OPTIONS);
  return sanitized;
};
