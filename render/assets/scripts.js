htmx.defineExtension("obj-enc", {
  onEvent: function (name, evt) {
    if (name === "htmx:configRequest") {
      evt.detail.headers["Content-Type"] = "application/json";
    }
  },

  encodeParameters: function (xhr, parameters, elt) {
    xhr.overrideMimeType("application/json");

    // If elt is not a form, find closest form
    if (elt.tagName !== "FORM") {
      elt = htmx.closest(elt, "form");
    }

    const formData = new FormData(elt);
    const obj = {};

    // Add missing checkboxes to the form data
    elt.querySelectorAll("input[type=checkbox]").forEach((val) => {
      if (val.name !== "" && !formData.has(val.name)) {
        formData.append(val.name, false);
      }
    });

    formData.forEach(function (v, key) {
      let out = v;

      // Infer the type from the corresponding input element (DOM)
      const inputElements = elt.querySelectorAll(`[name="${key}"]`);
      if (inputElements.length > 0) {
        const input = inputElements[0]; // First element
        switch (input.type) {
          case "number":
          case "range":
            out = v === "" ? null : Number(v);
            break;
          case "checkbox":
            out = input.checked;
            break;
          case "radio":
            out = v;
            if (v === "true" || v === "false") {
              out = v === "true";
            } else if (!isNaN(v) && v.trim() !== "") {
              out = Number(v);
            }
            break;
        }
      }

      // Handle multiple values (e.g., checkboxes with the same name)
      if (obj.hasOwnProperty(key)) {
        if (!Array.isArray(obj[key])) {
          obj[key] = [obj[key]];
        }
        obj[key].push(out);
      } else {
        obj[key] = out;
      }
    });

    return JSON.stringify(obj);
  },
});

function onInit() {
  const htmlElement = document.documentElement;

  if (
    localStorage.getItem("mode") === "dark" ||
    (!("mode" in localStorage) &&
      window.matchMedia("(prefers-color-scheme: dark)").matches)
  ) {
    htmlElement.classList.add("dark");
  } else {
    htmlElement.classList.remove("dark");
  }

  htmlElement.classList.add(localStorage.getItem("theme") || "uk-theme-zinc");
}
onInit();

function removeClosest(target, selector) {
  const li = target.closest(selector);
  if (li) {
    li.remove();
  }
}
