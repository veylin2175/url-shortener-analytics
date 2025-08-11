async function shortenUrl() {
    const urlInput = document.getElementById('url').value;
    const aliasInput = document.getElementById('alias').value;
    const resultDiv = document.getElementById('result');

    if (!urlInput) {
        resultDiv.innerHTML = '<p style="color: red;">Пожалуйста, введите URL.</p>';
        return;
    }

    const payload = {
        url: urlInput,
    };
    if (aliasInput) {
        payload.alias = aliasInput;
    }

    try {
        const response = await fetch('/shorten', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload),
        });

        const data = await response.json();

        if (response.ok) {
            const shortUrl = `${window.location.origin}/s/${data.alias}`;
            resultDiv.innerHTML = `<p>Сокращенная ссылка: <a href="${shortUrl}" target="_blank">${shortUrl}</a></p>`;
        } else {
            resultDiv.innerHTML = `<p style="color: red;">Ошибка: ${data.error}</p>`;
        }
    } catch (error) {
        resultDiv.innerHTML = `<p style="color: red;">Произошла ошибка при отправке запроса.</p>`;
        console.error('Error:', error);
    }
}

// Функция для получения аналитики
async function getAnalytics() {
    const aliasInput = document.getElementById('analytics-alias').value;
    const analyticsResultDiv = document.getElementById('analytics-result');

    if (!aliasInput) {
        analyticsResultDiv.innerHTML = '<p style="color: red;">Пожалуйста, введите псевдоним.</p>';
        return;
    }

    try {
        const response = await fetch(`/analytics/${aliasInput}`, {
            method: 'GET',
        });

        const data = await response.json();

        if (response.ok) {
            analyticsResultDiv.innerHTML = `
                <h3>Аналитика для "${aliasInput}"</h3>
                <pre>${JSON.stringify(data, null, 2)}</pre>
            `;
        } else {
            analyticsResultDiv.innerHTML = `<p style="color: red;">Ошибка: ${data.error}</p>`;
        }
    } catch (error) {
        analyticsResultDiv.innerHTML = `<p style="color: red;">Произошла ошибка при отправке запроса.</p>`;
        console.error('Error:', error);
    }
}