/**
 * design.js — FinFin Backoffice UI Animations
 * Uses Motion (vanilla JS Framer Motion) from CDN.
 * Zero changes to app logic — pure visual layer.
 */

import { animate, stagger, spring } from 'https://cdn.jsdelivr.net/npm/motion@11/+esm';

/* ── ENTRANCE ANIMATIONS ─────────────────────────────────────────────── */
function initEntrance() {
  animate(
    'header',
    { opacity: [0, 1], y: [-18, 0] },
    { duration: 0.45, easing: [0.25, 1, 0.5, 1] }
  );

  animate(
    '.kpi-card',
    { opacity: [0, 1], y: [16, 0], scale: [0.96, 1] },
    { duration: 0.4, delay: stagger(0.055, { start: 0.15 }), easing: [0.25, 1, 0.5, 1] }
  );

  animate(
    '.card',
    { opacity: [0, 1], y: [14, 0] },
    { duration: 0.4, delay: stagger(0.06, { start: 0.25 }), easing: [0.25, 1, 0.5, 1] }
  );
}

/* ── KPI COUNTER FLASH ───────────────────────────────────────────────── */
function initKpiAnimations() {
  document.querySelectorAll('.kpi-card strong').forEach((el) => {
    new MutationObserver(() => {
      animate(
        el,
        { scale: [1.2, 1], color: ['#a5b4fc', '#f8fafc'] },
        { duration: 0.35, easing: spring({ stiffness: 380, damping: 22 }) }
      );
    }).observe(el, { childList: true, characterData: true, subtree: true });
  });
}

/* ── BUTTON PRESS FEEDBACK ───────────────────────────────────────────── */
function initButtonAnimations() {
  document.querySelectorAll('button').forEach((btn) => {
    btn.addEventListener('pointerdown', () => {
      animate(btn, { scale: [1, 0.95] }, { duration: 0.1 });
    });
    btn.addEventListener('pointerup', () => {
      animate(btn, { scale: [0.95, 1] }, { duration: 0.2, easing: spring({ stiffness: 500, damping: 26 }) });
    });
    btn.addEventListener('pointerleave', () => {
      animate(btn, { scale: 1 }, { duration: 0.15 });
    });
  });
}

/* ── STATUS BADGE TRANSITION ─────────────────────────────────────────── */
function initBadgeAnimations() {
  document.querySelectorAll('.status').forEach((badge) => {
    new MutationObserver(() => {
      animate(
        badge,
        { scale: [0.88, 1.06, 1], opacity: [0.6, 1] },
        { duration: 0.3, easing: spring({ stiffness: 420, damping: 28 }) }
      );
    }).observe(badge, {
      attributes: true,
      attributeFilter: ['class'],
      childList: true,
      characterData: true,
      subtree: true,
    });
  });
}

/* ── TIMELINE STEP ANIMATION ─────────────────────────────────────────── */
function initTimelineAnimations() {
  const observer = new MutationObserver((mutations) => {
    mutations.forEach((m) => {
      const li = m.target.nodeType === 1 ? m.target : m.target.parentElement?.closest('li');
      if (!li) return;
      animate(
        li,
        { x: [-10, 0], opacity: [0.5, 1] },
        { duration: 0.28, easing: [0.25, 1, 0.5, 1] }
      );
    });
  });

  document.querySelectorAll('#stateTimeline li').forEach((li) =>
    observer.observe(li, { childList: true, characterData: true, subtree: true, attributes: true })
  );
}

/* ── AUDIT ROW ENTRANCE ──────────────────────────────────────────────── */
function initAuditAnimations() {
  const auditRows = document.getElementById('auditRows');
  if (!auditRows) return;

  new MutationObserver((mutations) => {
    mutations.forEach((m) => {
      m.addedNodes.forEach((node) => {
        if (node.nodeType === 1 && node.classList?.contains('audit-row')) {
          animate(
            node,
            { opacity: [0, 1], x: [-10, 0] },
            { duration: 0.22, easing: [0.25, 1, 0.5, 1] }
          );
        }
      });
    });
  }).observe(auditRows, { childList: true });
}

/* ── SPOTLIGHT EFFECT ON CARDS ───────────────────────────────────────── */
function initSpotlight() {
  document.querySelectorAll('.card').forEach((card) => {
    card.addEventListener('mousemove', (e) => {
      const r = card.getBoundingClientRect();
      card.style.setProperty('--mx', `${e.clientX - r.left}px`);
      card.style.setProperty('--my', `${e.clientY - r.top}px`);
    });
  });
}

/* ── LOADING SHIMMER ON PRE ──────────────────────────────────────────── */
function initLoadingState() {
  const statusBadge = document.getElementById('statusBadge');
  const pre = document.getElementById('responseJson');
  if (!statusBadge || !pre) return;

  new MutationObserver(() => {
    const sending = statusBadge.textContent?.startsWith('sending');
    pre.classList.toggle('loading', !!sending);
  }).observe(statusBadge, { childList: true, characterData: true, subtree: true });
}

/* ── SCENARIO CHIP CLICK BURST ───────────────────────────────────────── */
function initChipAnimations() {
  const chips = document.getElementById('scenarioChips');
  if (!chips) return;

  new MutationObserver(() => {
    chips.querySelectorAll('.scenario-chip').forEach((chip) => {
      if (chip._designBound) return;
      chip._designBound = true;
      chip.addEventListener('click', () => {
        animate(
          chip,
          { scale: [1, 0.93, 1.04, 1] },
          { duration: 0.3, easing: spring({ stiffness: 450, damping: 24 }) }
        );
      });
    });
  }).observe(chips, { childList: true });
}

/* ── INIT ────────────────────────────────────────────────────────────── */
document.addEventListener('DOMContentLoaded', () => {
  initEntrance();
  initKpiAnimations();
  initButtonAnimations();
  initBadgeAnimations();
  initTimelineAnimations();
  initAuditAnimations();
  initSpotlight();
  initLoadingState();
  initChipAnimations();
});
