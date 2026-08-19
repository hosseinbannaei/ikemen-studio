<script lang="ts">
  import { onMount } from 'svelte';
  import { projectStore } from './lib/stores/projectStore';
  import { engineStore } from './lib/stores/engineStore';
  import { settingsStore } from './lib/stores/settingsStore';

  import Sidebar from './lib/components/Sidebar.svelte';
  import Breadcrumb from './lib/components/Breadcrumb.svelte';
  import ProjectsView from './lib/components/ProjectsView.svelte';
  import ProjectWorkspaceView from './lib/components/ProjectWorkspaceView.svelte';
  import EngineManagerView from './lib/components/EngineManagerView.svelte';
  import SettingsView from './lib/components/SettingsView.svelte';

  import NewProjectModal from './lib/components/NewProjectModal.svelte';
  import OpenProjectModal from './lib/components/OpenProjectModal.svelte';
  import Toast from './lib/components/Toast.svelte';

  let activeTab: 'projects' | 'engines' | 'settings' = 'projects';
  let projectsSubView: 'list' | 'workspace' = 'list';

  let showNewProjectModal = false;
  let showOpenProjectModal = false;

  onMount(async () => {
    await settingsStore.load();
    await engineStore.loadInstalled();
    await projectStore.loadRecent();

    // If a project is already loaded, default to workspace
    if ($projectStore.current) {
      projectsSubView = 'workspace';
    }
  });

  // Switch to workspace when a project is loaded
  $: if ($projectStore.current && projectsSubView === 'list' && activeTab === 'projects') {
    // Keep user in list if they intentionally navigated back, but allow easy switching
  }

  function handleSelectTab(tab: 'projects' | 'engines' | 'settings') {
    activeTab = tab;
    if (tab === 'projects' && $projectStore.current) {
      // If user clicks Projects tab while on another tab, return to workspace or list
    }
  }

  function handleOpenProjectWorkspace() {
    activeTab = 'projects';
    projectsSubView = 'workspace';
  }

  function handleBackToProjectsList() {
    projectsSubView = 'list';
  }

  $: breadcrumbItems = (() => {
    if (activeTab === 'projects') {
      if (projectsSubView === 'workspace' && $projectStore.current) {
        return [
          { label: 'Projects', onClick: () => (projectsSubView = 'list') },
          { label: $projectStore.current.name },
        ];
      }
      return [{ label: 'Projects' }];
    } else if (activeTab === 'engines') {
      return [{ label: 'Engines' }];
    } else {
      return [{ label: 'Settings' }];
    }
  })();
</script>

<div class="flex h-screen w-screen bg-dark-900 text-slate-100 overflow-hidden select-none font-sans">
  <!-- Left Sidebar (Unity Hub Style) -->
  <Sidebar
    {activeTab}
    onSelectTab={handleSelectTab}
    onOpenProjectWorkspace={handleOpenProjectWorkspace}
  />

  <!-- Main View Area -->
  <div class="flex-1 flex flex-col min-w-0 overflow-hidden bg-dark-900">
    <!-- Unified Top Breadcrumb Header -->
    <Breadcrumb
      items={breadcrumbItems}
      onBack={activeTab === 'projects' && projectsSubView === 'workspace' ? handleBackToProjectsList : null}
      backLabel="Projects"
      showPlayButton={activeTab === 'projects' && projectsSubView === 'workspace'}
    />

    <!-- Main Content Tab Router -->
    <main class="flex-1 overflow-y-auto">
      {#if activeTab === 'projects'}
        {#if projectsSubView === 'workspace' && $projectStore.current}
          <ProjectWorkspaceView onBackToProjects={handleBackToProjectsList} />
        {:else}
          <ProjectsView
            onNewProject={() => (showNewProjectModal = true)}
            onOpenProjectWorkspace={handleOpenProjectWorkspace}
          />
        {/if}
      {:else if activeTab === 'engines'}
        <EngineManagerView />
      {:else if activeTab === 'settings'}
        <SettingsView />
      {/if}
    </main>
  </div>

  <!-- Modals -->
  {#if showNewProjectModal}
    <NewProjectModal
      onClose={() => (showNewProjectModal = false)}
      onOpenEngines={() => {
        showNewProjectModal = false;
        activeTab = 'engines';
      }}
    />
  {/if}

  {#if showOpenProjectModal}
    <OpenProjectModal onClose={() => (showOpenProjectModal = false)} />
  {/if}

  <!-- Toast Notification Overlay -->
  <Toast />
</div>
