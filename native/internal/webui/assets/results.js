(() => {
  "use strict";

  const dialogs = Array.from(document.querySelectorAll("[data-image-dialog]"));
  if (!dialogs.length) return;

  const byId = new Map(dialogs.map((dialog) => [dialog.id, dialog]));

  function openerFor(dialog) {
    const id = dialog.dataset.openerId;
    return id ? document.getElementById(id) : null;
  }

  function focusClose(dialog) {
    const close = dialog.querySelector("[data-image-close]");
    if (close instanceof HTMLElement) close.focus();
  }

  function openDialog(dialog) {
    if (!(dialog instanceof HTMLDialogElement)) return;
    if (typeof dialog.showModal === "function") {
      if (!dialog.open) dialog.showModal();
    } else {
      dialog.setAttribute("open", "");
    }
    focusClose(dialog);
  }

  function closeDialog(dialog, restoreFocus = true) {
    if (!(dialog instanceof HTMLDialogElement)) return;
    dialog.dataset.restoreFocus = restoreFocus ? "true" : "false";
    if (dialog.open && typeof dialog.close === "function") dialog.close();
    else dialog.removeAttribute("open");
    if (restoreFocus) {
      const opener = openerFor(dialog);
      if (opener instanceof HTMLElement) opener.focus();
    }
  }

  function switchDialog(current, offset) {
    const index = dialogs.indexOf(current);
    if (index < 0 || dialogs.length < 2) return;
    const next = dialogs[(index + offset + dialogs.length) % dialogs.length];
    if (next === current) return;
    closeDialog(current, false);
    openDialog(next);
  }

  document.addEventListener("click", (event) => {
    const openControl = event.target.closest("[data-image-open]");
    if (openControl) {
      const dialog = byId.get(openControl.dataset.dialog || "");
      if (dialog) {
        event.preventDefault();
        openDialog(dialog);
      }
      return;
    }

    const closeControl = event.target.closest("[data-image-close]");
    if (closeControl) {
      const dialog = closeControl.closest("[data-image-dialog]");
      if (dialog) closeDialog(dialog, true);
      return;
    }

    const previousControl = event.target.closest("[data-image-previous]");
    if (previousControl) {
      const dialog = previousControl.closest("[data-image-dialog]");
      if (dialog) switchDialog(dialog, -1);
      return;
    }

    const nextControl = event.target.closest("[data-image-next]");
    if (nextControl) {
      const dialog = nextControl.closest("[data-image-dialog]");
      if (dialog) switchDialog(dialog, 1);
    }
  });

  for (const dialog of dialogs) {
    dialog.addEventListener("click", (event) => {
      if (event.target === dialog) closeDialog(dialog, true);
    });
    dialog.addEventListener("keydown", (event) => {
      if (event.key === "ArrowLeft") {
        event.preventDefault();
        switchDialog(dialog, -1);
      } else if (event.key === "ArrowRight") {
        event.preventDefault();
        switchDialog(dialog, 1);
      }
    });
    dialog.addEventListener("close", () => {
      if (dialog.dataset.restoreFocus !== "false") {
        const opener = openerFor(dialog);
        if (opener instanceof HTMLElement) opener.focus();
      }
      delete dialog.dataset.restoreFocus;
    });
  }
})();
