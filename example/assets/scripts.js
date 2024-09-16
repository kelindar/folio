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

document.addEventListener("DOMContentLoaded", function () {
  var comboboxes = document.querySelectorAll(".combobox");

  comboboxes.forEach(function (combobox) {
    var comboboxInput = combobox.querySelector(".combobox-input");
    var comboboxToggle = combobox.querySelector(".combobox-toggle");
    var optionsList = combobox.querySelector(".combobox-options");
    var options = combobox.querySelectorAll(".combobox-option");
    var hiddenInput = combobox.querySelector(".combobox-value");

    // Filter options as the user types
    comboboxInput.addEventListener("input", function () {
      var filter = comboboxInput.value.toLowerCase();
      var anyVisible = false;

      options.forEach(function (option) {
        var text = option.textContent.toLowerCase();
        if (text.includes(filter)) {
          option.style.display = "";
          anyVisible = true;
        } else {
          option.style.display = "none";
        }
      });

      // Open the dropdown if any options are visible
      comboboxToggle.checked = anyVisible;
    });

    // Update input value and hidden input when an option is clicked
    options.forEach(function (option) {
      option.addEventListener("click", function () {
        var selectedValue = option.getAttribute("data-value");
        comboboxInput.value = option.textContent;
        hiddenInput.value = selectedValue;
        comboboxToggle.checked = false;

        // Optionally, trigger any additional events or validation here
      });
    });

    // Close the dropdown when clicking outside
    document.addEventListener("click", function (event) {
      if (!combobox.contains(event.target)) {
        comboboxToggle.checked = false;
      }
    });
  });
});
