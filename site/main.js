/* markmaton — Blueprint Engine Room
   One orchestrated moment: the conversion theater.
   Everything else stays quiet. */

(() => {
  const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  /* ---------- conversion theater ---------- */

  const theater = document.querySelector(".theater");
  const noiseGroups = Array.from(theater.querySelectorAll(".noise"))
    .sort((a, b) => Number(a.dataset.n) - Number(b.dataset.n));

  const STRIKE_EVERY = 240; // ms between strike groups
  const PRINT_AFTER = 380; // pause before markdown prints

  let timers = [];

  const clearTimers = () => {
    timers.forEach(clearTimeout);
    timers = [];
  };

  const playTheater = () => {
    clearTimers();
    theater.classList.remove("played");
    noiseGroups.forEach((g) => g.classList.remove("struck"));

    if (reduced) {
      noiseGroups.forEach((g) => g.classList.add("struck"));
      theater.classList.add("played");
      return;
    }

    noiseGroups.forEach((group, i) => {
      timers.push(setTimeout(() => group.classList.add("struck"), 500 + i * STRIKE_EVERY));
    });
    timers.push(
      setTimeout(() => theater.classList.add("played"), 500 + noiseGroups.length * STRIKE_EVERY + PRINT_AFTER)
    );
  };

  // Start when the figure first scrolls into view; replay on demand.
  const startOnce = new IntersectionObserver(
    (entries) => {
      if (entries.some((e) => e.isIntersecting)) {
        playTheater();
        startOnce.disconnect();
      }
    },
    { threshold: 0.35 }
  );
  startOnce.observe(theater);

  document.querySelector(".replay").addEventListener("click", playTheater);

  /* ---------- copy buttons ---------- */

  document.querySelectorAll(".copy").forEach((btn) => {
    btn.addEventListener("click", async () => {
      const text = btn.dataset.copy;
      try {
        await navigator.clipboard.writeText(text);
      } catch {
        const area = document.createElement("textarea");
        area.value = text;
        document.body.appendChild(area);
        area.select();
        document.execCommand("copy");
        area.remove();
      }
      btn.classList.add("copied");
      setTimeout(() => btn.classList.remove("copied"), 1400);
    });
  });

  /* ---------- quiet scroll reveals for sheets ---------- */

  if (!reduced) {
    const targets = document.querySelectorAll(".sheet-label, .sheet h2, .sheet-lede, .pipeline, .callout, .output-grid, .stat-row, .install-grid, .compare-rows, .dim-rule");
    targets.forEach((el) => el.classList.add("reveal"));
    const revealOnView = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add("visible");
            revealOnView.unobserve(entry.target);
          }
        });
      },
      { threshold: 0.2, rootMargin: "0px 0px -40px 0px" }
    );
    targets.forEach((el) => revealOnView.observe(el));
  }
})();
