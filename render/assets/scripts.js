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

// Loop through all dropdowns & comboboxes and hide them when clicking outside
document.addEventListener("click", function (event) {
  document
    .querySelectorAll(".btn-dropdown, .combobox")
    .forEach(function (element) {
      var toggle = element.querySelector(
        ".btn-dropdown-toggle, .combobox-toggle"
      );
      if (!element.contains(event.target)) {
        toggle.checked = false;
      }
    });
});

// Initialize the combobox on load
function onComboboxLoad(event) {
  var cb = event.target;
  var hi = cb.querySelector(".combobox-value");
  var cbi = cb.querySelector(".combobox-input");
  var opts = cb.querySelectorAll(".combobox-option");
  var val = hi.value;

  opts.forEach(function (opt) {
    if (opt.getAttribute("data-value") === val) {
      opt.classList.add("selected");
      cbi.value = opt.querySelector(".block.truncate").textContent.trim();
    } else {
      opt.classList.remove("selected");
    }
  });
}

// Handle input events to filter options
function onComboboxInput(event) {
  var cbi = event.target;
  var cb = cbi.closest(".combobox");
  var opts = cb.querySelectorAll(".combobox-option");
  var cbt = cb.querySelector(".combobox-toggle");
  var filter = cbi.value.toLowerCase();
  var anyVisible = false;

  opts.forEach(function (opt) {
    var text = opt.querySelector(".block.truncate").textContent.toLowerCase();
    var match = text.includes(filter);
    opt.style.display = match ? "" : "none";
    anyVisible = anyVisible || match;
  });

  cbt.checked = anyVisible;
}

// Handle keydown events on the combobox input
function onComboboxKeyDown(event) {
  var cbi = event.target;
  if (event.key === "Enter" || event.keyCode === 13) {
    event.preventDefault(); // Prevent form submission
    selectFirstVisibleOption(cbi);
  }
}

// Handle click events on options
function onComboboxOptionClick(event) {
  var opt = event.currentTarget;
  selectOption(opt);
}

// Helper function to select an option
function selectOption(opt) {
  var cb = opt.closest(".combobox");
  var cbi = cb.querySelector(".combobox-input");
  var hi = cb.querySelector(".combobox-value");
  var cbt = cb.querySelector(".combobox-toggle");

  cb.querySelectorAll(".combobox-option").forEach(function (o) {
    o.classList.remove("selected");
  });

  opt.classList.add("selected");
  cbi.value = opt.querySelector(".block.truncate").textContent.trim();
  hi.value = opt.getAttribute("data-value");
  cbt.checked = false;
}

// Helper function to select the first visible option
function selectFirstVisibleOption(cbi) {
  var cb = cbi.closest(".combobox");
  var opts = cb.querySelectorAll(".combobox-option");
  for (var i = 0; i < opts.length; i++) {
    var opt = opts[i];
    if (opt.style.display !== "none") {
      selectOption(opt);
      break;
    }
  }
}
