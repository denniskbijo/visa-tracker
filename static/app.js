(function () {
  'use strict';

  var STORAGE_KEY = 'visa-tracker-timeline';

  // --- Wizard ---
  function initWizard() {
    var form = document.getElementById('wizard-form');
    if (!form) return;

    var steps = form.querySelectorAll('.wizard-step');
    var prevBtn = document.getElementById('wizard-prev');
    var nextBtn = document.getElementById('wizard-next');
    var submitBtn = document.getElementById('wizard-submit');
    var stepNum = document.getElementById('wizard-step-num');
    var progress = document.getElementById('wizard-progress');
    var current = 1;
    var total = steps.length;

    function showStep(n) {
      current = n;
      steps.forEach(function (step) {
        step.classList.toggle('hidden', parseInt(step.dataset.step, 10) !== n);
      });
      if (stepNum) stepNum.textContent = String(n);
      if (progress) progress.style.width = ((n / total) * 100) + '%';
      if (prevBtn) prevBtn.classList.toggle('hidden', n === 1);
      if (nextBtn) nextBtn.classList.toggle('hidden', n === total);
      if (submitBtn) submitBtn.classList.toggle('hidden', n !== total);
    }

    function stepValid(n) {
      var step = form.querySelector('.wizard-step[data-step="' + n + '"]');
      if (!step) return true;
      var radios = step.querySelectorAll('input[type="radio"]');
      if (radios.length > 0) {
        return Array.prototype.some.call(radios, function (r) { return r.checked; });
      }
      return true;
    }

    if (nextBtn) {
      nextBtn.addEventListener('click', function () {
        if (!stepValid(current)) {
          alert('Please select an option to continue.');
          return;
        }
        if (current < total) showStep(current + 1);
      });
    }

    if (prevBtn) {
      prevBtn.addEventListener('click', function () {
        if (current > 1) showStep(current - 1);
      });
    }

    showStep(1);
  }

  // --- Timeline ---
  function initTimeline() {
    var form = document.getElementById('timeline-form');
    var dataEl = document.getElementById('timeline-routes-data');
    if (!form || !dataEl) return;

    var routes;
    try {
      routes = JSON.parse(dataEl.textContent);
    } catch (e) {
      return;
    }

    var routeBySlug = {};
    routes.forEach(function (r) { routeBySlug[r.slug] = r; });

    function load() {
      try {
        return JSON.parse(localStorage.getItem(STORAGE_KEY));
      } catch (e) {
        return null;
      }
    }

    function save(data) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
    }

    function clear() {
      localStorage.removeItem(STORAGE_KEY);
    }

    function addYears(date, years) {
      var d = new Date(date);
      d.setFullYear(d.getFullYear() + Math.floor(years));
      var months = Math.round((years % 1) * 12);
      if (months) d.setMonth(d.getMonth() + months);
      return d;
    }

    function daysBetween(from, to) {
      var ms = to.getTime() - from.getTime();
      return Math.ceil(ms / (1000 * 60 * 60 * 24));
    }

    function fmtDate(d) {
      return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'long', year: 'numeric' });
    }

    function render(data) {
      var empty = document.getElementById('timeline-empty');
      var stats = document.getElementById('timeline-stats');
      if (!data || !data.route || !data.start) {
        if (empty) empty.classList.remove('hidden');
        if (stats) stats.classList.add('hidden');
        return;
      }

      var route = routeBySlug[data.route];
      if (!route) return;

      if (empty) empty.classList.add('hidden');
      if (stats) stats.classList.remove('hidden');

      var start = new Date(data.start + 'T00:00:00');
      var today = new Date();
      today.setHours(0, 0, 0, 0);

      var expiry = addYears(start, route.durationYears);
      var daysLeft = daysBetween(today, expiry);

      var daysEl = document.getElementById('stat-days-left');
      var expiryEl = document.getElementById('stat-expiry');
      if (daysEl) {
        daysEl.textContent = daysLeft > 0 ? String(daysLeft) : '0';
        daysEl.classList.toggle('warn', daysLeft > 0 && daysLeft <= 90);
        daysEl.classList.toggle('urgent', daysLeft <= 0);
      }
      if (expiryEl) {
        expiryEl.textContent = daysLeft <= 0
          ? 'Expired on ' + fmtDate(expiry)
          : 'Expires ' + fmtDate(expiry);
      }

      var ilrBlock = document.getElementById('stat-ilr-block');
      var ilrDaysEl = document.getElementById('stat-ilr-days');
      var ilrDateEl = document.getElementById('stat-ilr-date');
      var noteBox = document.getElementById('stat-note-box');
      var noteEl = document.getElementById('stat-note');

      if (route.ilrYears > 0) {
        if (ilrBlock) ilrBlock.classList.remove('hidden');
        var ilrDate = addYears(start, route.ilrYears);
        var ilrDays = daysBetween(today, ilrDate);
        if (ilrDaysEl) ilrDaysEl.textContent = ilrDays > 0 ? String(ilrDays) : '0';
        if (ilrDateEl) {
          ilrDateEl.textContent = ilrDays <= 0
            ? 'Eligible since ' + fmtDate(ilrDate)
            : 'Eligible from ' + fmtDate(ilrDate);
        }
      } else {
        if (ilrBlock) ilrBlock.classList.add('hidden');
      }

      if (route.note && noteBox && noteEl) {
        noteBox.style.display = 'block';
        noteEl.textContent = route.note;
      } else if (noteBox) {
        noteBox.style.display = 'none';
      }

      var link = document.getElementById('stat-route-link');
      var nameEl = document.getElementById('stat-route-name');
      if (link) link.href = '/visas/' + route.slug;
      if (nameEl) nameEl.textContent = data.label || route.name;
    }

    function populateForm(data) {
      if (!data) return;
      var routeSel = document.getElementById('timeline-route');
      var startIn = document.getElementById('timeline-start');
      var labelIn = document.getElementById('timeline-label');
      if (routeSel && data.route) routeSel.value = data.route;
      if (startIn && data.start) startIn.value = data.start;
      if (labelIn && data.label) labelIn.value = data.label;
    }

    form.addEventListener('submit', function (e) {
      e.preventDefault();
      var data = {
        route: document.getElementById('timeline-route').value,
        start: document.getElementById('timeline-start').value,
        label: document.getElementById('timeline-label').value.trim()
      };
      save(data);
      render(data);
    });

    var clearBtn = document.getElementById('timeline-clear');
    if (clearBtn) {
      clearBtn.addEventListener('click', function () {
        clear();
        form.reset();
        render(null);
      });
    }

    var saved = load();
    populateForm(saved);
    render(saved);
  }

  document.addEventListener('DOMContentLoaded', function () {
    initWizard();
    initTimeline();
  });
})();
