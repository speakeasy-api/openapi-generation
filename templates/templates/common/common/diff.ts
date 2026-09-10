"use strict";

function diff(left: any, right: any): { changed: boolean; text: string } {
  const defaultOptions = {
    indent: "  ",
    newLine: "\n",
    wrap: function wrap(type: "modified" | "added" | "removed", text: string) {
      return chalk(type, text);
    },
    color: true,
  };
  function chalk(type: "modified" | "added" | "removed", text: string): string {
    const colors = {
      modified: [33, 39], // yellow
      added: [32, 39], // green
      removed: [31, 39], // red
    };
    return `\x1B[${colors[type]}m${text}\x1B[${colors[type]}m`;
  }

  function diffInternal(
    left: any,
    right: any,
    options,
  ): { changed: boolean; text: string } {
    function isObject(obj: any) {
      return typeof obj === "object" && obj && !Array.isArray(obj);
    }

    function printVar(variable: any) {
      if (typeof variable === "function") {
        return variable.toString().replace(/\{.+\}/, "{}");
      } else if (
        (typeof variable === "object" || typeof variable === "string") &&
        !(variable instanceof RegExp)
      ) {
        try {
          return JSON.stringify(variable);
        } catch (e) {
          return "[Circular]";
        }
      }

      return "" + variable;
    }

    function indentSubItem(text, options) {
      return text
        .split(options.newLine)
        .map(function onMap(line, index) {
          if (index === 0) {
            return line;
          }
          return options.indent + line;
        })
        .join(options.newLine);
    }

    function keyChanged(key, text, options) {
      return (
        options.indent +
        key +
        ": " +
        indentSubItem(text, options) +
        options.newLine
      );
    }

    function keyRemoved(key, variable, options) {
      return (
        options.wrap("removed", "- " + key + ": " + printVar(variable)) +
        options.newLine
      );
    }

    function keyAdded(key, variable, options) {
      return (
        options.wrap("added", "+ " + key + ": " + printVar(variable)) +
        options.newLine
      );
    }

    let text = "";
    let changed = false;
    let itemDiff;
    let keys;
    let subOutput = "";

    if (Array.isArray(left) && Array.isArray(right)) {
      for (var i = 0; i < left.length; i++) {
        if (i < right.length) {
          itemDiff = diffInternal(left[i], right[i], options);
          if (itemDiff.changed) {
            subOutput += keyChanged(i, itemDiff.text, options);
            changed = true;
          }
        } else {
          subOutput += keyRemoved(i, left[i], options);
          changed = true;
        }
      }
      if (right.length > left.length) {
        for (; i < right.length; i++) {
          subOutput += keyAdded(i, right[i], options);
        }
        changed = true;
      }
      if (changed) {
        text = "[" + options.newLine + subOutput + "]";
      }
    } else if (isObject(left) && isObject(right)) {
      keys = Object.keys(left);
      var rightObj = Object.assign({}, right);
      var key;
      keys.sort();
      for (let i = 0; i < keys.length; i++) {
        key = keys[i];
        if (right.hasOwnProperty(key)) {
          itemDiff = diffInternal(left[key], right[key], options);
          if (itemDiff.changed) {
            subOutput += keyChanged(key, itemDiff.text, options);
            changed = true;
          }
          delete rightObj[key];
        } else {
          subOutput += keyRemoved(key, left[key], options);
          changed = true;
        }
      }

      let addedKeys = Object.keys(rightObj);
      for (let i = 0; i < addedKeys.length; i++) {
        subOutput += keyAdded(addedKeys[i], right[addedKeys[i]], options);
        changed = true;
      }

      if (changed) {
        text = "{" + options.newLine + subOutput + "}";
      }
    } else if (typeof left == "function" && typeof right == "function") {
      changed = false;
    } else if (left !== right) {
      text = options.wrap(
        "modified",
        printVar(left) + " => " + printVar(right),
      );
      changed = true;
    }

    return {
      changed: changed,
      text: text,
    };
  }

  return diffInternal(left, right, defaultOptions);
}
