/**
 * Credits to https://gist.github.com/colinrubbert/230d757384d92606ed8bc267e9158ed3
 * RuntimeGlobalsChecker
 *
 * You can use this utility to quickly check what variables have been added (or
 * leaked) to the global window object at runtime (by JavaScript code).
 * By running this code, the globals checker itself is attached as a singleton
 * to the window object as "__runtimeGlobalsChecker__".
 * You can check the runtime globals programmatically at any time by invoking
 * "window.__runtimeGlobalsChecker__.getRuntimeGlobals()".
 *
 */
window.__runtimeGlobalsChecker__ = (function createGlobalsChecker() {
  // Globals on the window object set by default by the browser.
  // We collect them to then filter them out of from the list of globals (since
  // we don't care about them).
  // They're populated by "collectBrowserGlobals()" and will contain globals such
  // as "location" and "localStorage".
  let browserGlobals = [];
 
  // Known globals on the window object that we can safely ignored.
  // This list should be populated manually after trial and errors.
  const ignoredGlobals = ["__runtimeGlobalsChecker__"];
 
  /**
   * Collect the global variables added to the window object by the browser by
   * creating a temporary iframe and checking what global variables the browser
   * adds on it.
   * @returns {string[]} - List of globals added added by the browser
   */
  function collectBrowserGlobals() {
    const iframe = window.document.createElement("iframe");
    iframe.src = "about:blank";
    window.document.body.appendChild(iframe);
    browserGlobals = Object.keys(iframe.contentWindow);
    window.document.body.removeChild(iframe);
    return browserGlobals;
  }
 
  /**
   * Return the list of globals added at runtime (by JavaScript).
   * @returns {string[]} - List of globals added at runtime (by JavaScript)
   */
  function getRuntimeGlobals() {
    // If we haven't collected the browser globals yet, do it now.
    if (browserGlobals.length === 0) {
      collectBrowserGlobals();
    }
    // Grab all the globals filtering out variables we don't care about (noise).
    const runtimeGlobals = Object.keys(window).filter((key) => {
      const isFromBrowser = browserGlobals.includes(key);
      const isIgnored = ignoredGlobals.includes(key);
      return !isFromBrowser && !isIgnored;
    });
    return runtimeGlobals;
  }
 
  return {
    getRuntimeGlobals,
  };
})();
