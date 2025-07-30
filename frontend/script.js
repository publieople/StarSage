document.addEventListener('DOMContentLoaded', () => {
    const repoList = document.getElementById('repo-list');
    const searchBox = document.getElementById('search-box');
    let allRepos = [];

    function renderRepos(repos) {
        repoList.innerHTML = '';
        if (repos.length === 0) {
            repoList.innerHTML = '<p>No repositories found.</p>';
            return;
        }

        repos.forEach(repo => {
            const repoItem = document.createElement('div');
            repoItem.className = 'repo-item';

            const description = repo.Description ? `<p>${repo.Description}</p>` : '';
            const language = repo.Language ? `<p class="language">Language: ${repo.Language}</p>` : '';

            repoItem.innerHTML = `
                <h2><a href="${repo.URL}" target="_blank">${repo.FullName}</a></h2>
                ${description}
                ${language}
                <p>⭐ ${repo.StargazersCount}</p>
            `;
            repoList.appendChild(repoItem);
        });
    }

    function filterRepos() {
        const query = searchBox.value.toLowerCase();
        const filteredRepos = allRepos.filter(repo => {
            return repo.FullName.toLowerCase().includes(query) ||
                   (repo.Description && repo.Description.toLowerCase().includes(query));
        });
        renderRepos(filteredRepos);
    }

    async function fetchRepos() {
        try {
            const response = await fetch('/api/repositories');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            allRepos = await response.json();
            renderRepos(allRepos);
        } catch (error) {
            repoList.innerHTML = `<p>Error loading repositories: ${error.message}</p>`;
        }
    }

    searchBox.addEventListener('input', filterRepos);
    fetchRepos();
});
