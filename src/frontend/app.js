// Configuração do HTMX
document.body.addEventListener('htmx:configRequest', (evt) => {
    // Adiciona o token de autenticação se existir
    const token = localStorage.getItem('token');
    if (token) {
        evt.detail.headers['Authorization'] = `Bearer ${token}`;
    }
});

// Manipulador para respostas de autenticação
document.body.addEventListener('htmx:afterRequest', (evt) => {
    const response = JSON.parse(evt.detail.xhr.response);
    const target = document.getElementById('auth-response') || document.getElementById('form-response');

    if (target) {
        if (evt.detail.xhr.status >= 200 && evt.detail.xhr.status < 300) {
            if (response.token) {
                localStorage.setItem('token', response.token);
                target.innerHTML = '<div class="success">Operação realizada com sucesso! Redirecionando...</div>';
                setTimeout(() => {
                    window.location.href = 'index.html';
                }, 1500);
            } else {
                target.innerHTML = '<div class="success">Operação realizada com sucesso!</div>';
            }
        } else {
            target.innerHTML = `<div class="error">${response.detail || 'Erro ao processar a requisição'}</div>`;
        }
    }
});

// Manipulador para exibir o modal
document.body.addEventListener('htmx:afterSwap', (evt) => {
    if (evt.detail.target.id === 'auth-modal') {
        document.getElementById('auth-modal').classList.add('active');
    }
});

// Fechar modal ao clicar fora
document.getElementById('auth-modal').addEventListener('click', (evt) => {
    if (evt.target === evt.currentTarget) {
        evt.currentTarget.classList.remove('active');
    }
});

// Template para exibir produtos
function productTemplate(product) {
    return `
        <div class="product-card">
            <img src="${product.image_url || 'https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRyHCW1Uw5t2nTLwa6BhMx4RFE2heuXSox3Yw&s'}" 
                 alt="${product.description}" 
                 class="product-image">
            <div class="product-info">
                <h2 class="product-title">${product.description}</h2>
                <p class="product-price">R$ ${product.value.toFixed(2)}</p>
                <p class="product-quantity">Quantidade: ${product.quantity}</p>
            </div>
        </div>
    `;
}

// Manipulador para a resposta da lista de produtos
document.body.addEventListener('htmx:afterSwap', (evt) => {
    if (evt.detail.target.id === 'products') {
        const products = JSON.parse(evt.detail.xhr.response);
        if (Array.isArray(products)) {
            evt.detail.target.innerHTML = products.map(productTemplate).join('');
        }
    }
});

// Verificar autenticação ao carregar a página
document.addEventListener('DOMContentLoaded', () => {
    const token = localStorage.getItem('token');
    const nav = document.querySelector('nav .container');

    if (nav) {
        if (token) {
            // Usuário autenticado
            const navLinks = document.createElement('div');
            navLinks.className = 'nav-links';
            navLinks.innerHTML = `
                <a href="index.html">Home</a>
                <a href="create-product.html">Criar Produto</a>
                <button onclick="logout()">Sair</button>
            `;
            nav.appendChild(navLinks);
        } else {
            // Usuário não autenticado
            const authButtons = document.createElement('div');
            authButtons.className = 'nav-links';
            authButtons.innerHTML = `
                <a href="login.html">Login</a>
                <a href="register.html">Registrar</a>
            `;
            nav.appendChild(authButtons);
        }
    }
});

// Função de logout
function logout() {
    localStorage.removeItem('token');
    window.location.href = 'index.html';
} 