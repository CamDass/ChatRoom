// Confirmation avant déconnexion
document.querySelectorAll('a[href="/logout"]').forEach(el => {
    el.addEventListener('click', e => {
        if (!confirm('Se déconnecter ?')) e.preventDefault()
    })
})

// Auto-resize textarea
document.querySelectorAll('textarea').forEach(el => {
    el.addEventListener('input', function () {
        this.style.height = 'auto'
        this.style.height = this.scrollHeight + 'px'
    })
})



// ── quote ──────────────────────────────────────────────────────────────

function quotePost(id, username, content, hasImage) {
    const imageTag = hasImage ? ' [image]' : ''
    const preview = content.length > 120 
        ? content.substring(0, 120) + '...' + imageTag
        : content + imageTag

    document.getElementById('parent-id-input').value = id
    document.getElementById('quote-author').textContent = '@' + username
    document.getElementById('quote-text').textContent = preview
    document.getElementById('quote-preview').style.display = 'block'
    document.querySelector('.reply-box textarea').focus()
    document.querySelector('.reply-box').scrollIntoView({ behavior: 'smooth' })
}

function clearQuote() {
    document.getElementById('parent-id-input').value = ''
    document.getElementById('quote-preview').style.display = 'none'
}




// ── Modal profil ──────────────────────────────────────────────────────────────
// Déclenché uniquement sur les .mention-link présents dans topic.html
const overlay  = document.getElementById('profile-modal-overlay')
const modalBox = document.getElementById('profile-modal-box')
const closeBtn = document.getElementById('profile-modal-close')

function openProfileModal(username) {
    if (!overlay) return
    const content = document.getElementById('profile-modal-content')
    content.innerHTML = '<div class="profile-modal-loading">Chargement…</div>'
    overlay.style.display = 'flex'

    fetch('/user/' + encodeURIComponent(username) + '?modal=1')
        .then(r => r.text())
        .then(html => {
            // Extraire uniquement le bloc .profile-page du HTML retourné
            const tmp = document.createElement('div')
            tmp.innerHTML = html
            const profilePage = tmp.querySelector('.profile-page')
            if (profilePage) {
                content.innerHTML = ''
                content.appendChild(profilePage)
            } else {
                content.innerHTML = '<p style="color:var(--muted);text-align:center">Profil introuvable.</p>'
            }
        })
        .catch(() => {
            content.innerHTML = '<p style="color:var(--danger);text-align:center">Erreur de chargement.</p>'
        })
}

function closeProfileModal() {
    if (overlay) overlay.style.display = 'none'
}

// Clic sur un @mention dans topic.html
document.querySelectorAll('.mention-link').forEach(el => {
    el.style.cursor = 'pointer'
    el.addEventListener('click', e => {
        e.preventDefault()
        // Clic droit ou Ctrl+clic → tag dans la textarea
        if (e.ctrlKey || e.metaKey) {
            tagUser(el.dataset.username)
        } else {
            openProfileModal(el.dataset.username)
        }
    })
})

// Fermeture
if (closeBtn) closeBtn.addEventListener('click', closeProfileModal)
if (overlay) overlay.addEventListener('click', e => {
    if (e.target === overlay) closeProfileModal()
})
document.addEventListener('keydown', e => {
    if (e.key === 'Escape') closeProfileModal()
})



// ── Tag @mention dans la textarea ─────────────────────────────
function tagUser(username) {
    const textarea = document.getElementById('reply-content')
    if (!textarea) return
    const tag = '@' + username + ' '
    const pos = textarea.selectionStart
    textarea.value = textarea.value.slice(0, pos) + tag + textarea.value.slice(pos)
    textarea.focus()
    textarea.selectionStart = textarea.selectionEnd = pos + tag.length
    textarea.scrollIntoView({ behavior: 'smooth' })
}

// ── Autocomplétion @mention ───────────────────────────────────
function initMentionAutocomplete(textarea) {
    if (!textarea) return

    const dropdown = document.createElement('div')
    dropdown.className = 'mention-dropdown'
    dropdown.style.display = 'none'
    textarea.parentNode.style.position = 'relative'
    textarea.parentNode.appendChild(dropdown)

    let currentQuery = null
    let selectedIndex = -1
    let currentUsers = []

    function showDropdown(users) {
        currentUsers = users
        selectedIndex = -1
        dropdown.innerHTML = ''
        if (!users.length) { dropdown.style.display = 'none'; return }
        users.forEach((u, i) => {
            const item = document.createElement('div')
            item.className = 'mention-item'
            item.textContent = '@' + u
            item.addEventListener('mousedown', e => {
                e.preventDefault()
                insertMention(u)
            })
            dropdown.appendChild(item)
        })
        dropdown.style.display = 'block'
    }

    function hideDropdown() {
        dropdown.style.display = 'none'
        currentUsers = []
        selectedIndex = -1
        currentQuery = null
    }

    function insertMention(username) {
        const val = textarea.value
        const pos = textarea.selectionStart
        const before = val.slice(0, pos)
        const atIndex = before.lastIndexOf('@')
        textarea.value = val.slice(0, atIndex) + '@' + username + ' ' + val.slice(pos)
        const newPos = atIndex + username.length + 2
        textarea.selectionStart = textarea.selectionEnd = newPos
        hideDropdown()
        textarea.focus()
    }

    function highlightItem(index) {
        const items = dropdown.querySelectorAll('.mention-item')
        items.forEach((el, i) => el.classList.toggle('active', i === index))
    }

    textarea.addEventListener('input', () => {
        const val = textarea.value
        const pos = textarea.selectionStart
        const before = val.slice(0, pos)
        const match = before.match(/@(\w*)$/)

        if (!match) { hideDropdown(); return }

        const query = match[1]
        if (query === currentQuery) return
        currentQuery = query

        if (query.length === 0) { hideDropdown(); return }

        fetch('/api/users?q=' + encodeURIComponent(query))
            .then(r => r.json())
            .then(users => showDropdown(users || []))
            .catch(() => hideDropdown())
    })

    textarea.addEventListener('keydown', e => {
        if (dropdown.style.display === 'none') return
        if (e.key === 'ArrowDown') {
            e.preventDefault()
            selectedIndex = Math.min(selectedIndex + 1, currentUsers.length - 1)
            highlightItem(selectedIndex)
        } else if (e.key === 'ArrowUp') {
            e.preventDefault()
            selectedIndex = Math.max(selectedIndex - 1, 0)
            highlightItem(selectedIndex)
        } else if (e.key === 'Enter' || e.key === 'Tab') {
            if (selectedIndex >= 0) {
                e.preventDefault()
                insertMention(currentUsers[selectedIndex])
            }
        } else if (e.key === 'Escape') {
            hideDropdown()
        }
    })

    document.addEventListener('click', e => {
        if (!dropdown.contains(e.target) && e.target !== textarea) hideDropdown()
    })
}

// Init sur la textarea de réponse
initMentionAutocomplete(document.getElementById('reply-content'))