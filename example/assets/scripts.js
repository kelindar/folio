htmx.defineExtension("obj-enc", {
  onEvent: function (name, evt) {
    if (name === "htmx:configRequest") {
      evt.detail.headers["Content-Type"] = "application/json";
    }
  },

  encodeParameters: function (xhr, parameters, elt) {
    xhr.overrideMimeType("text/json");

    const formData = new FormData(elt);
    const obj = {};

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
