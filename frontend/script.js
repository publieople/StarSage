document.addEventListener('DOMContentLoaded', () => {
    // State
    let state = {
        currentView: 'repositories', // 'repositories' or 'lists'
        allRepos: [],
        allLists: [],
    };

    // DOM Elements
    const nav = {
        repos: document.getElementById('nav-repos'),
        lists: document.getElementById('nav-lists'),
    };
    const views = {
        repositories: document.getElementById('repositories-view'),
        lists: document.getElementById('lists-view'),
    };
    const repoListContainer = document.getElementById('repo-list');
    const listContainer = document.getElementById('ai-lists-container');
    const searchBox = document.getElementById('search-box');

    // Modal Elements
    const modal = document.getElementById('create-list-modal');
    const createListBtn = document.getElementById('create-list-btn');
    const closeModalBtn = modal.querySelector('.close-btn');
    const createListForm = document.getElementById('create-list-form');
    const listNameInput = document.getElementById('list-name');
    const listPromptInput = document.getElementById('list-prompt');

    // --- RENDER FUNCTIONS ---

    function renderRepos(repos) {
        repoListContainer.innerHTML = '';
        if (repos.length === 0) {
            repoListContainer.innerHTML = '<p>No repositories found.</p>';
            return;
        }

        repos.forEach(repo => {
            const repoItem = document.createElement('div');
            repoItem.className = 'repo-item';
            const description = repo.Description ? `<p>${repo.Description}</p>` : '';
            const language = repo.Language ? `<p class="language">Language: ${repo.Language}</p>` : '';
            const summary = repo.Summary ? `<p><strong>AI Summary:</strong> ${repo.Summary}</p>` : '';

            repoItem.innerHTML = `
                <h2><a href="${repo.URL}" target="_blank">${repo.FullName}</a></h2>
                ${description}
                ${summary}
                ${language}
                <p>⭐ ${repo.StargazersCount}</p>
            `;
            repoListContainer.appendChild(repoItem);
        });
    }

    function renderLists(lists) {
        listContainer.innerHTML = '';
        if (lists.length === 0) {
            listContainer.innerHTML = '<p>No AI-generated lists found. Create one!</p>';
            return;
        }

        lists.forEach(list => {
            const listItem = document.createElement('div');
            listItem.className = 'repo-item'; // Reuse the same style
            listItem.innerHTML = `
                <h2>${list.Name}</h2>
                <p><em>${list.Prompt}</em></p>
                <p>${list.RepoCount} repositories</p>
            `;
            // TODO: Add click handler to view repos in the list
            listContainer.appendChild(listItem);
        });
    }

    // --- API FUNCTIONS ---

    async function fetchRepos() {
        try {
            const response = await fetch('/api/repositories');
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            state.allRepos = await response.json();
            renderRepos(state.allRepos);
        } catch (error) {
            repoListContainer.innerHTML = `<p>Error loading repositories: ${error.message}</p>`;
        }
    }

    async function fetchLists() {
        try {
            const response = await fetch('/api/lists');
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            state.allLists = await response.json();
            renderLists(state.allLists);
        } catch (error) {
            listContainer.innerHTML = `<p>Error loading lists: ${error.message}</p>`;
        }
    }

    async function createList(name, prompt) {
        try {
            const response = await fetch('/api/lists', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name, prompt }),
            });
            const result = await response.json();
            if (!response.ok) {
                throw new Error(result.error || `HTTP error! status: ${response.status}`);
            }
            alert('List creation started! It will appear in the list shortly.');
            closeModal();
            fetchLists(); // Refresh the list view
        } catch (error) {
            alert(`Error creating list: ${error.message}`);
        }
    }

    // --- VIEW & MODAL MANAGEMENT ---

    function showView(viewName) {
        state.currentView = viewName;
        Object.keys(views).forEach(key => {
            views[key].classList.toggle('hidden', key !== viewName);
        });
        Object.keys(nav).forEach(key => {
            nav[key].classList.toggle('active', key.startsWith(viewName.slice(0, 4)));
        });

        if (viewName === 'lists') {
            fetchLists();
        }
    }

    function openModal() {
        modal.classList.remove('hidden');
    }

    function closeModal() {
        modal.classList.add('hidden');
        createListForm.reset();
    }

    // --- EVENT LISTENERS ---

    function filterRepos() {
        const query = searchBox.value.toLowerCase();
        const filteredRepos = state.allRepos.filter(repo => {
            return repo.FullName.toLowerCase().includes(query) ||
                   (repo.Description && repo.Description.toLowerCase().includes(query)) ||
                   (repo.Summary && repo.Summary.toLowerCase().includes(query));
        });
        renderRepos(filteredRepos);
    }

    nav.repos.addEventListener('click', (e) => {
        e.preventDefault();
        showView('repositories');
    });

    nav.lists.addEventListener('click', (e) => {
        e.preventDefault();
        showView('lists');
    });

    searchBox.addEventListener('input', filterRepos);

    createListBtn.addEventListener('click', openModal);
    closeModalBtn.addEventListener('click', closeModal);
    window.addEventListener('click', (e) => {
        if (e.target === modal) closeModal();
    });

    createListForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const name = listNameInput.value.trim();
        const prompt = listPromptInput.value.trim();
        if (name && prompt) {
            createList(name, prompt);
        }
    });

    // --- INITIALIZATION ---

    function init() {
        showView('repositories');
        fetchRepos();
    }

    init();
});
