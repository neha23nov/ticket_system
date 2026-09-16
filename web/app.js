// ==========================================================================
// Ticket System — Frontend Application Logic
//
// This file handles all the client-side functionality:
//   - User authentication (login, register, logout)
//   - Ticket management (create, list, view, update status)
//   - Page navigation (switching between views)
//   - API communication with the backend
//
// The JWT token is stored in localStorage so the user stays logged in
// even after refreshing the page (until they log out or the token expires).
// ==========================================================================

// --------------------------------------------------------------------------
// Global State
// --------------------------------------------------------------------------

// allTickets stores the full list of the user's tickets, fetched from the API.
// This allows filtering by status without making extra API calls.
let allTickets = [];

// currentFilter tracks which status filter is currently active on the dashboard.
let currentFilter = 'all';

// --------------------------------------------------------------------------
// API Helper
// --------------------------------------------------------------------------

/**
 * Makes an HTTP request to the backend API.
 *
 * @param {string} method  - HTTP method (GET, POST, PATCH, etc.)
 * @param {string} path    - API path (e.g., "/auth/login", "/tickets")
 * @param {object} body    - Request body (will be sent as JSON). Pass null for GET.
 * @param {boolean} auth   - Whether to include the JWT token in the request.
 * @returns {object}       - { ok: boolean, status: number, data: object }
 */
async function api(method, path, body = null, auth = true) {
    const headers = { 'Content-Type': 'application/json' };

    // If authentication is required, add the JWT token from localStorage.
    if (auth) {
        const token = localStorage.getItem('token');
        if (token) {
            headers['Authorization'] = 'Bearer ' + token;
        }
    }

    const options = { method, headers };
    if (body) {
        options.body = JSON.stringify(body);
    }

    try {
        const response = await fetch(path, options);
        const data = await response.json();
        return { ok: response.ok, status: response.status, data };
    } catch (error) {
        return { ok: false, status: 0, data: { error: 'Network error. Is the server running?' } };
    }
}

// --------------------------------------------------------------------------
// Page Navigation
// --------------------------------------------------------------------------

/**
 * Shows a specific page and hides all others.
 * Pages are identified by their HTML element ID.
 *
 * @param {string} pageId - The ID of the page element to show.
 */
function showPage(pageId) {
    // Hide all pages.
    document.querySelectorAll('.page').forEach(p => p.classList.add('hidden'));

    // Show the requested page.
    document.getElementById(pageId).classList.remove('hidden');
}

/**
 * Shows the login/register page and hides the navigation bar.
 */
function showAuth() {
    document.getElementById('navbar').classList.add('hidden');
    showPage('auth-page');
}

/**
 * Shows the dashboard (ticket list) and loads tickets from the API.
 */
async function showDashboard() {
    showPage('dashboard-page');
    await loadTickets();
}

/**
 * Shows the "Create New Ticket" form.
 */
function showCreateTicket() {
    showPage('create-page');
    // Clear any previous form data.
    document.getElementById('ticket-title').value = '';
    document.getElementById('ticket-desc').value = '';
}

/**
 * Shows the detail view for a specific ticket.
 *
 * @param {number} ticketId - The ID of the ticket to display.
 */
async function showTicketDetail(ticketId) {
    showPage('detail-page');

    const result = await api('GET', '/tickets/' + ticketId);
    if (!result.ok) {
        showToast(result.data.error || 'Failed to load ticket', 'error');
        showDashboard();
        return;
    }

    renderTicketDetail(result.data);
}

// --------------------------------------------------------------------------
// Authentication
// --------------------------------------------------------------------------

/**
 * Switches between the Login and Register tabs on the auth page.
 *
 * @param {string} tab - Either 'login' or 'register'.
 */
function switchAuthTab(tab) {
    // Update tab styling.
    document.getElementById('tab-login').classList.toggle('active', tab === 'login');
    document.getElementById('tab-register').classList.toggle('active', tab === 'register');

    // Show/hide the corresponding form.
    document.getElementById('login-form').classList.toggle('hidden', tab !== 'login');
    document.getElementById('register-form').classList.toggle('hidden', tab !== 'register');

    // Clear any messages.
    hideAuthMessage();
}

/**
 * Handles the login form submission.
 * Sends credentials to the API and stores the JWT token on success.
 */
async function handleLogin(event) {
    event.preventDefault();

    const email = document.getElementById('login-email').value.trim();
    const password = document.getElementById('login-password').value;

    const result = await api('POST', '/auth/login', { email, password }, false);

    if (result.ok) {
        // Store the token and username.
        localStorage.setItem('token', result.data.token);
        localStorage.setItem('email', email);

        // Clear form and show dashboard.
        document.getElementById('login-form').reset();
        onLoginSuccess();
    } else {
        showAuthMessage(result.data.error || 'Login failed', 'error');
    }
}

/**
 * Handles the registration form submission.
 * Creates a new account and switches to the login tab on success.
 */
async function handleRegister(event) {
    event.preventDefault();

    const username = document.getElementById('reg-username').value.trim();
    const email = document.getElementById('reg-email').value.trim();
    const password = document.getElementById('reg-password').value;

    const result = await api('POST', '/auth/register', { username, email, password }, false);

    if (result.ok) {
        showAuthMessage('Account created! Please log in.', 'success');
        document.getElementById('register-form').reset();

        // Auto-fill the login form with the registered email.
        document.getElementById('login-email').value = email;

        // Switch to login tab after a short delay.
        setTimeout(() => switchAuthTab('login'), 1500);
    } else {
        showAuthMessage(result.data.error || 'Registration failed', 'error');
    }
}

/**
 * Logs the user out by clearing stored data and returning to the auth page.
 */
function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('email');
    localStorage.removeItem('username');
    allTickets = [];
    showAuth();
}

/**
 * Called after a successful login. Sets up the UI for an authenticated user.
 */
function onLoginSuccess() {
    const email = localStorage.getItem('email') || 'User';
    document.getElementById('nav-username').textContent = email;
    document.getElementById('navbar').classList.remove('hidden');
    showDashboard();
}

// --------------------------------------------------------------------------
// Auth Messages
// --------------------------------------------------------------------------

function showAuthMessage(text, type) {
    const el = document.getElementById('auth-message');
    el.textContent = text;
    el.className = 'message message-' + type;
    el.classList.remove('hidden');
}

function hideAuthMessage() {
    document.getElementById('auth-message').classList.add('hidden');
}

// --------------------------------------------------------------------------
// Tickets
// --------------------------------------------------------------------------

/**
 * Loads all tickets for the current user from the API and renders them.
 */
async function loadTickets() {
    const result = await api('GET', '/tickets');

    if (!result.ok) {
        if (result.status === 401) {
            // Token expired or invalid — redirect to login.
            logout();
            return;
        }
        showToast('Failed to load tickets', 'error');
        return;
    }

    allTickets = result.data || [];
    renderTickets();
}

/**
 * Handles the "Create Ticket" form submission.
 */
async function handleCreateTicket(event) {
    event.preventDefault();

    const title = document.getElementById('ticket-title').value.trim();
    const description = document.getElementById('ticket-desc').value.trim();

    const result = await api('POST', '/tickets', { title, description });

    if (result.ok) {
        showToast('Ticket created successfully!', 'success');
        showDashboard();
    } else {
        showToast(result.data.error || 'Failed to create ticket', 'error');
    }
}

/**
 * Updates a ticket's status.
 *
 * @param {number} ticketId - The ID of the ticket to update.
 * @param {string} newStatus - The new status ("in_progress" or "closed").
 */
async function updateTicketStatus(ticketId, newStatus) {
    const result = await api('PATCH', '/tickets/' + ticketId + '/status', { status: newStatus });

    if (result.ok) {
        showToast('Status updated to ' + formatStatus(newStatus), 'success');
        renderTicketDetail(result.data);
        // Also refresh the ticket in our local cache.
        const idx = allTickets.findIndex(t => t.id === ticketId);
        if (idx !== -1) allTickets[idx] = result.data;
    } else {
        showToast(result.data.error || 'Failed to update status', 'error');
    }
}

/**
 * Filters the ticket list by status.
 *
 * @param {string} filter - The status to filter by, or 'all' for no filter.
 * @param {HTMLElement} btn - The clicked filter button (for styling).
 */
function filterTickets(filter, btn) {
    currentFilter = filter;

    // Update active tab styling.
    document.querySelectorAll('.filter-tab').forEach(t => t.classList.remove('active'));
    btn.classList.add('active');

    renderTickets();
}

// --------------------------------------------------------------------------
// Rendering
// --------------------------------------------------------------------------

/**
 * Renders the ticket list on the dashboard, applying the current filter.
 */
function renderTickets() {
    const list = document.getElementById('ticket-list');
    const empty = document.getElementById('empty-state');

    // Apply filter.
    const filtered = currentFilter === 'all'
        ? allTickets
        : allTickets.filter(t => t.status === currentFilter);

    if (filtered.length === 0) {
        list.classList.add('hidden');
        empty.classList.remove('hidden');
    } else {
        list.classList.remove('hidden');
        empty.classList.add('hidden');

        list.innerHTML = filtered.map(ticket => `
            <div class="ticket-card" onclick="showTicketDetail(${ticket.id})">
                <div class="ticket-info">
                    <div class="ticket-title">${escapeHtml(ticket.title)}</div>
                    <div class="ticket-meta">
                        #${ticket.id} · Created ${formatDate(ticket.created_at)}
                    </div>
                </div>
                <span class="status-badge status-${ticket.status}">
                    ${formatStatus(ticket.status)}
                </span>
            </div>
        `).join('');
    }
}

/**
 * Renders the ticket detail view, including status update buttons.
 *
 * @param {object} ticket - The ticket object from the API.
 */
function renderTicketDetail(ticket) {
    const container = document.getElementById('ticket-detail');

    // Determine which status transitions are available.
    let statusActions = '';
    if (ticket.status === 'open') {
        statusActions = `
            <button class="btn btn-sm btn-primary" onclick="updateTicketStatus(${ticket.id}, 'in_progress')">
                → Move to In Progress
            </button>`;
    } else if (ticket.status === 'in_progress') {
        statusActions = `
            <button class="btn btn-sm btn-success" onclick="updateTicketStatus(${ticket.id}, 'closed')">
                ✓ Close Ticket
            </button>`;
    }

    container.innerHTML = `
        <div class="detail-header">
            <h2 class="detail-title">${escapeHtml(ticket.title)}</h2>
            <span class="status-badge status-${ticket.status}">
                ${formatStatus(ticket.status)}
            </span>
        </div>

        <div class="detail-section">
            <div class="detail-label">Description</div>
            <div class="detail-description">${escapeHtml(ticket.description)}</div>
        </div>

        <div class="detail-meta">
            <div>
                <div class="detail-label">Ticket ID</div>
                <div>#${ticket.id}</div>
            </div>
            <div>
                <div class="detail-label">Created</div>
                <div>${formatDate(ticket.created_at)}</div>
            </div>
            <div>
                <div class="detail-label">Last Updated</div>
                <div>${formatDate(ticket.updated_at)}</div>
            </div>
            <div>
                <div class="detail-label">Status</div>
                <div>${formatStatus(ticket.status)}</div>
            </div>
        </div>

        ${statusActions ? `
        <div class="status-update-section">
            <h4>Update Status</h4>
            <div class="status-actions">${statusActions}</div>
        </div>` : `
        <div class="status-update-section">
            <p style="color: var(--color-text-light); font-size: 0.9rem;">
                ✅ This ticket is closed. No further status changes are allowed.
            </p>
        </div>`}
    `;
}

// --------------------------------------------------------------------------
// Utility Functions
// --------------------------------------------------------------------------

/**
 * Converts a status string to a human-readable label.
 * e.g., "in_progress" becomes "In Progress"
 */
function formatStatus(status) {
    const labels = {
        'open': 'Open',
        'in_progress': 'In Progress',
        'closed': 'Closed'
    };
    return labels[status] || status;
}

/**
 * Formats an ISO date string into a readable format.
 * e.g., "2024-01-15T10:30:00Z" becomes "Jan 15, 2024"
 */
function formatDate(dateStr) {
    if (!dateStr) return 'N/A';
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

/**
 * Escapes HTML special characters to prevent XSS attacks.
 * This is used whenever we display user-provided content.
 */
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * Shows a toast notification at the bottom-right of the screen.
 *
 * @param {string} message - The message to display.
 * @param {string} type    - 'success' or 'error'.
 */
function showToast(message, type) {
    const toast = document.getElementById('toast');
    toast.textContent = message;
    toast.className = 'toast toast-' + type;
    toast.classList.remove('hidden');

    // Auto-hide the toast after 3 seconds.
    setTimeout(() => {
        toast.classList.add('hidden');
    }, 3000);
}

// --------------------------------------------------------------------------
// App Initialization
// --------------------------------------------------------------------------

/**
 * Runs when the page first loads.
 * Checks if the user has a stored JWT token and either shows the
 * dashboard (if logged in) or the auth page (if not).
 */
(function init() {
    const token = localStorage.getItem('token');
    if (token) {
        // User has a stored token — try to load the dashboard.
        onLoginSuccess();
    } else {
        // No token — show the login page.
        showAuth();
    }
})();
