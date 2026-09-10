String.prototype.replaceAll = function (str, newStr) {
  // If a regex pattern
  if (Object.prototype.toString.call(str).toLowerCase() === "[object regexp]") {
    return this.replace(str, newStr);
  }

  // If a string
  return this.replace(new RegExp(escapeRegExp(str), "g"), newStr);
};

function escapeRegExp(string) {
  return string.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"); // $& means the whole matched string
}

// Allow debug in templates
registerTemplateFunc("debug", debug);

if (!console.time || !console.timeEnd) {
  const timers = {};

  console.time = function (label = "default") {
    timers[label] = Date.now();
  };

  console.timeEnd = function (label = "default") {
    if (!timers[label]) {
      console.log(`No such label: ${label}`);
      return;
    }
    const duration = Date.now() - timers[label];
    console.log(`${label}: ${duration}ms`);
    delete timers[label];
  };
}
