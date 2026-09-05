(() => {
  "use strict";

  const storageKey = "goreecloud.search.preferences.v1";
  const schemaVersion = 1;
  const defaultTheme = "system";
  const defaultDensity = "comfortable";
  const themes = new Set(["system", "light", "dark", "deep-dark"]);
  const densityProfiles = Object.freeze({
    comfortable: "comfortable",
    compact: "productive",
  });

  function read() {
    try {
      const raw = localStorage.getItem(storageKey);
      if (!raw) return { theme: defaultTheme, density: defaultDensity };
      const envelope = JSON.parse(raw);
      if (!envelope || envelope.schema_version !== schemaVersion || !envelope.preferences || typeof envelope.preferences !== "object") {
        return { theme: defaultTheme, density: defaultDensity };
      }
      const theme = themes.has(envelope.preferences["appearance.theme"])
        ? envelope.preferences["appearance.theme"]
        : defaultTheme;
      const density = Object.hasOwn(densityProfiles, envelope.preferences["appearance.result_density"])
        ? envelope.preferences["appearance.result_density"]
        : defaultDensity;
      return { theme, density };
    } catch (_error) {
      return { theme: defaultTheme, density: defaultDensity };
    }
  }

  function apply(theme = defaultTheme, density = defaultDensity) {
    const root = document.documentElement;
    root.dataset.glazeVersion = "1.1";

    if (themes.has(theme) && theme !== "system") {
      root.dataset.glzAppearance = theme;
    } else {
      delete root.dataset.glzAppearance;
    }

    root.dataset.glazeDensityProfile = densityProfiles[density] || densityProfiles[defaultDensity];
  }

  const current = read();
  apply(current.theme, current.density);

  window.GoreeCloudSearchAppearance = Object.freeze({ apply, read });
})();
