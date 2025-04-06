const API_BASE_URL = 'http://localhost:8000/api/v1';

async function fetchAPI(endpoint, options = {}) {
    const token = localStorage.getItem('token');
    const headers = {
        'Content-Type': 'application/json',
        ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
        ...options.headers
    };

    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            ...options,
            headers
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.detail || 'Erro na requisição');
        }

        return data;
    } catch (error) {
        console.error('API Error:', error);
        throw error;
    }
}

export const authAPI = {
    async login(email, password) {
        return fetchAPI('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        });
    },

    async register(name, email, password) {
        return fetchAPI('/auth/register', {
            method: 'POST',
            body: JSON.stringify({ name, email, password })
        });
    }
};

export const productAPI = {
    async list() {
        const products = await fetchAPI('/products');
        // Adiciona a URL completa da imagem usando a rota correta
        return products.map(product => ({
            ...product,
            image_url: product.image ? `${API_BASE_URL}/images/${product.image}` : null
        }));
    },

    async create(productData) {
        const formData = new FormData();
        formData.append('description', productData.description);
        formData.append('value', productData.value.toString());
        formData.append('quantity', productData.quantity.toString());
        if (productData.image) {
            formData.append('image', productData.image);
        }

        const token = localStorage.getItem('token');
        const response = await fetch(`${API_BASE_URL}/products`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },
            body: formData
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.detail || 'Erro ao criar produto');
        }

        const data = await response.json();
        return {
            ...data,
            image_url: data.image ? `${API_BASE_URL}/images/${data.image}` : null
        };
    }
};

export function showMessage(element, message, type = 'error') {
    element.innerHTML = `
        <div class="message ${type}">
            ${message}
        </div>
    `;
    element.style.display = 'block';

    if (type === 'success') {
        setTimeout(() => {
            element.style.display = 'none';
        }, 3000);
    }
} 