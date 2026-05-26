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