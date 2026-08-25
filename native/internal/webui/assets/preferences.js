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
    localStorage.setItem(storageKey, JSON.stringify({
      schema_version: schemaVersion,
      preferences: sanitize(preferences),
    }));
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
    });
  }

  function bindFilter() {
    const input = document.querySelector("[data-settings-filter]");
    if (!input) return;
    const rows = [...document.querySelectorAll(".setting-row")];
    const sections = [...document.querySelectorAll(".settings-section")];
    input.addEventListener("input", () => {
      const query = input.value.trim().toLocaleLowerCase();
      rows.forEach((row) => {
        row.hidden = query !== "" && !row.textContent.toLocaleLowerCase().includes(query);
      });
      sections.forEach((section) => {
        const visible = [...section.querySelectorAll(".setting-row")].some((row) => !row.hidden);
        section.hidden = query !== "" && !visible;
      });
    });
  }

  function downloadJSON(filename, value) {
    const blob = new Blob([JSON.stringify(value, null, 2) + "\n"], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    anchor.click();
    URL.revokeObjectURL(url);
  }

  function bindPortability() {
    document.querySelector("[data-export-preferences]")?.addEventListener("click", () => {
      downloadJSON("goreecloud-search-preferences.json", {
        product: "GoreeCloud Search",
        schema_version: schemaVersion,
        preferences: readPreferences(),
      });
    });

    document.querySelector("[data-reset-preferences]")?.addEventListener("click", () => {
      localStorage.removeItem(storageKey);
      location.reload();
    });

    const importInput = document.querySelector("[data-import-preferences]");
    importInput?.addEventListener("change", async () => {
      const [file] = importInput.files || [];
      if (!file || file.size > 64 * 1024) return;
      try {
        const envelope = JSON.parse(await file.text());
        if (envelope.schema_version !== schemaVersion || envelope.product !== "GoreeCloud Search") return;
        writePreferences(envelope.preferences);
        location.reload();
      } catch (_error) {
        // Invalid imports are intentionally ignored without exposing parser details.
      }
    });
  }

  const preferences = readPreferences();
  document.querySelectorAll("select[data-preference]").forEach((element) => bindSelect(element, preferences));
  document.querySelectorAll("button.toggle[data-preference]").forEach((element) => bindToggle(element, preferences));
  applyTheme(preferences["appearance.theme"]);
  bindFilter();
  bindPortability();
})();
