(() => {
  "use strict";

  const storageKey = "goreecloud.search.preferences.v1";
  const schemaVersion = 1;
  const defaults = {
    "search.default_category": "general",
    "search.safe_search": "moderate",
    "search.autocomplete": false,
    "appearance.theme": "system",
    "appearance.result_density": "comfortable",
    "privacy.recent_queries": false,
  };

  const allowed = {
    "search.default_category": ["general", "images", "videos", "news", "files"],
    "search.safe_search": ["off", "moderate", "strict"],
    "appearance.theme": ["system", "light", "dark"],
    "appearance.result_density": ["comfortable", "compact"],
  };

  const message = document.querySelector("[data-settings-message]");
  function announce(text, state = "success") {
    if (!message) return;
    message.textContent = text;
    message.dataset.state = state;
  }

  function sanitize(candidate) {
    const next = { ...defaults };
    if (!candidate || typeof candidate !== "object" || Array.isArray(candidate)) return next;
    for (const [key, value] of Object.entries(candidate)) {
      if (!(key in defaults)) continue;
      if (typeof defaults[key] === "boolean") {
        if (typeof value === "boolean") next[key] = value;
        continue;
      }
      if (allowed[key] && allowed[key].includes(value)) next[key] = value;
    }
    return next;
  }

  function readPreferences() {
    try {
      const raw = localStorage.getItem(storageKey);
      if (!raw) return { ...defaults };
      const envelope = JSON.parse(raw);
      if (envelope.schema_version !== schemaVersion) return { ...defaults };
      return sanitize(envelope.preferences);
    } catch (_error) {
      return { ...defaults };
    }
  }

  function writePreferences(preferences) {
    localStorage.setItem(storageKey, JSON.stringify({ schema_version: schemaVersion, preferences: sanitize(preferences) }));
  }

  function applyTheme(theme) {
    document.documentElement.dataset.theme = theme;
    if (theme === "system") delete document.documentElement.dataset.theme;
  }

  function bindSelect(element, preferences) {
    const key = element.dataset.preference;
    if (!key) return;
    element.value = String(preferences[key]);
    element.addEventListener("change", () => {
      const next = readPreferences();
      next[key] = element.value;
      writePreferences(next);
      if (key === "appearance.theme") applyTheme(element.value);
      announce("Preference saved on this device.");
    });
  }

  function bindToggle(element, preferences) {
    const key = element.dataset.preference;
    if (!key) return;
    const update = (value) => {
      element.setAttribute("aria-pressed", String(value));
      element.classList.toggle("is-on", value);
    };
    update(Boolean(preferences[key]));
    element.addEventListener("click", () => {
      const next = readPreferences();
      next[key] = !Boolean(next[key]);
      writePreferences(next);
      update(next[key]);
      announce("Preference saved on this device.");
    });
  }

  function bindFilter() {
    const input = document.querySelector("[data-settings-filter]");
    const empty = document.querySelector("[data-empty-filter]");
    if (!input) return;
    const rows = [...document.querySelectorAll(".setting-row")];
    const sections = [...document.querySelectorAll("[data-settings-section]")];
    input.addEventListener("input", () => {
      const query = input.value.trim().toLocaleLowerCase();
      rows.forEach((row) => { row.hidden = query !== "" && !row.textContent.toLocaleLowerCase().includes(query); });
      let visibleSections = 0;
      sections.forEach((section) => {
        const visible = [...section.querySelectorAll(".setting-row")].some((row) => !row.hidden);
        section.hidden = query !== "" && !visible;
        if (!section.hidden) visibleSections += 1;
      });
      if (empty) empty.hidden = query === "" || visibleSections !== 0;
    });
  }

  function bindSectionNavigation() {
    const links = [...document.querySelectorAll("[data-settings-nav] a")];
    if (!("IntersectionObserver" in window) || links.length === 0) return;
    const lookup = new Map(links.map((link) => [link.getAttribute("href")?.slice(1), link]));
    const observer = new IntersectionObserver((entries) => {
      const visible = entries.filter((entry) => entry.isIntersecting).sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];
      if (!visible) return;
      links.forEach((link) => link.classList.remove("active"));
      lookup.get(visible.target.id)?.classList.add("active");
    }, { rootMargin: "-15% 0px -65% 0px", threshold: [0, 0.25, 0.5] });
    document.querySelectorAll("[data-settings-section]").forEach((section) => observer.observe(section));
  }

  function downloadJSON(filename, value) {
    const blob = new Blob([JSON.stringify(value, null, 2) + "\n"], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    anchor.click();
    setTimeout(() => URL.revokeObjectURL(url), 0);
  }

  function bindPortability() {
    document.querySelector("[data-export-preferences]")?.addEventListener("click", () => {
      downloadJSON("goreecloud-search-preferences.json", { product: "GoreeCloud Search", schema_version: schemaVersion, preferences: readPreferences() });
      announce("Preferences exported without deployment credentials.");
    });

    const reset = document.querySelector("[data-reset-preferences]");
    let resetArmed = false;
    reset?.addEventListener("click", () => {
      if (!resetArmed) {
        resetArmed = true;
        reset.textContent = "Confirm reset";
        announce("Select Confirm reset to restore local defaults.", "error");
        setTimeout(() => {
          resetArmed = false;
          reset.textContent = "Reset";
        }, 5000);
        return;
      }
      localStorage.removeItem(storageKey);
      location.reload();
    });

    const importInput = document.querySelector("[data-import-preferences]");
    importInput?.addEventListener("change", async () => {
      const [file] = importInput.files || [];
      importInput.value = "";
      if (!file) return;
      if (file.size > 64 * 1024) {
        announce("Import rejected: preference files must be 64 KiB or smaller.", "error");
        return;
      }
      try {
        const envelope = JSON.parse(await file.text());
        if (envelope.schema_version !== schemaVersion || envelope.product !== "GoreeCloud Search") {
          announce("Import rejected: incompatible GoreeCloud Search preference file.", "error");
          return;
        }
        writePreferences(envelope.preferences);
        announce("Preferences imported. Reloading…");
        location.reload();
      } catch (_error) {
        announce("Import rejected: the file is not valid preference JSON.", "error");
      }
    });
  }

  const preferences = readPreferences();
  document.querySelectorAll("select[data-preference]").forEach((element) => bindSelect(element, preferences));
  document.querySelectorAll("button.toggle[data-preference]").forEach((element) => bindToggle(element, preferences));
  applyTheme(preferences["appearance.theme"]);
  bindFilter();
  bindSectionNavigation();
  bindPortability();
})();
