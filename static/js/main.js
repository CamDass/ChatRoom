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



function quotePost(id, username, content) {
    document.getElementById('parent-id-input').value = id
    document.getElementById('quote-author').textContent = '@' + username
    document.getElementById('quote-text').textContent = content.length > 120 
        ? content.substring(0, 120) + '...' 
        : content
    document.getElementById('quote-preview').style.display = 'block'
    document.querySelector('.reply-box textarea').focus()
    document.querySelector('.reply-box').scrollIntoView({ behavior: 'smooth' })
}

function clearQuote() {
    document.getElementById('parent-id-input').value = ''
    document.getElementById('quote-preview').style.display = 'none'
}