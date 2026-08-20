<script lang="ts">
  import { onMount } from 'svelte';
  import { 
    X, Save, FileText, Code2, Layers, Search, 
    Play, AlertCircle, Check, Copy, ExternalLink,
    Sparkles, ArrowRight, Zap, RefreshCw, Hash, Cpu
  } from 'lucide-svelte';
  import { ReadProjectFile, SaveProjectFile, InspectProjectFile } from '../../../wailsjs/go/main/App';
  import { toastStore } from '../stores/toastStore';
  import type { FileInspectionResult } from '../types';

  export let projectDir: string;
  export let filePath: string;
  export let onClose: () => void;

  let inspection: FileInspectionResult | null = null;
  let rawText = '';
  let originalText = '';
  let isLoading = true;
  let isSaving = false;
  let activeTab: 'editor' | 'inspector' = 'editor';
  let searchQuery = '';
  let selectedSection: string = '';

  $: isDirty = rawText !== originalText;

  onMount(async () => {
    await loadFileData();
  });

  async function loadFileData() {
    isLoading = true;
    try {
      const res = await InspectProjectFile(projectDir, filePath);
      inspection = res as any;
      rawText = res.raw_content;
      originalText = res.raw_content;
    } catch (err: any) {
      toastStore.error('Load Failed', err?.message || String(err));
    } finally {
      isLoading = false;
    }
  }

  async function handleSave() {
    if (!isDirty || isSaving) return;
    isSaving = true;
    try {
      await SaveProjectFile(projectDir, filePath, rawText);
      originalText = rawText;
      toastStore.success('File Saved', filePath);
      // Re-inspect in background to update structural metrics
      inspection = (await InspectProjectFile(projectDir, filePath)) as any;
    } catch (err: any) {
      toastStore.error('Save Failed', err?.message || String(err));
    } finally {
      isSaving = false;
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key === 's') {
      e.preventDefault();
      handleSave();
    } else if (e.key === 'Escape') {
      if (!isDirty || confirm('Discard unsaved changes?')) {
        onClose();
      }
    }
  }

  function jumpToSection(sectionName: string) {
    selectedSection = sectionName;
    const lines = rawText.split('\n');
    let targetIndex = -1;
    for (let i = 0; i < lines.length; i++) {
      if (lines[i].toLowerCase().includes(sectionName.toLowerCase())) {
        targetIndex = i;
        break;
      }
    }
    if (targetIndex !== -1) {
      const textarea = document.getElementById('code-editor-textarea') as HTMLTextAreaElement;
      if (textarea) {
        const linePos = lines.slice(0, targetIndex).join('\n').length + (targetIndex > 0 ? 1 : 0);
        textarea.focus();
        textarea.setSelectionRange(linePos, linePos + lines[targetIndex].length);
        const lineHeight = 20;
        textarea.scrollTop = targetIndex * lineHeight - 100;
      }
    }
  }

  function copyContent() {
    navigator.clipboard.writeText(rawText);
    toastStore.info('Copied', 'Content copied to clipboard');
  }

  function getCategoryColor(category: string) {
    switch (category) {
      case 'fighter': return 'text-amber-400 bg-amber-500/10 border-amber-500/30';
      case 'stage': return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/30';
      case 'motif': return 'text-indigo-400 bg-indigo-500/10 border-indigo-500/30';
      case 'lifebar': return 'text-rose-400 bg-rose-500/10 border-rose-500/30';
      case 'animation': return 'text-purple-400 bg-purple-500/10 border-purple-500/30';
      case 'commands': return 'text-cyan-400 bg-cyan-500/10 border-cyan-500/30';
      default: return 'text-slate-400 bg-slate-800 border-slate-700';
    }
  }
</script>

<svelte:window on:keydown={handleKeyDown} />

<div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-in fade-in duration-200">
  <div class="relative w-full max-w-6xl h-[90vh] bg-dark-900 border border-dark-700 rounded-2xl shadow-2xl flex flex-col overflow-hidden text-slate-100 font-sans">
    
    <!-- Top Header Bar -->
    <div class="flex items-center justify-between px-6 py-3.5 border-b border-dark-700/80 bg-dark-950/60">
      <div class="flex items-center gap-3 min-w-0">
        <div class="w-9 h-9 rounded-xl bg-brand-500/10 border border-brand-500/30 flex items-center justify-center text-brand-400 flex-shrink-0">
          <Code2 class="w-5 h-5" />
        </div>
        <div class="min-w-0">
          <div class="flex items-center gap-2">
            <h2 class="text-sm font-bold text-white truncate font-mono">{filePath}</h2>
            {#if isDirty}
              <span class="px-2 py-0.5 text-[10px] font-bold rounded-full bg-amber-500/20 text-amber-300 border border-amber-500/30 animate-pulse">
                Unsaved
              </span>
            {/if}
            {#if inspection}
              <span class="px-2 py-0.5 text-[10px] font-bold uppercase rounded-full border {getCategoryColor(inspection.category)}">
                {inspection.category}
              </span>
            {/if}
          </div>
          <p class="text-xs text-slate-400 truncate">
            {#if inspection}
              {inspection.total_lines} lines • {(inspection.size_bytes / 1024).toFixed(1)} KB • Syntax: {inspection.syntax_mode.toUpperCase()}
            {:else}
              Loading file inspection...
            {/if}
          </p>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center gap-2">
        <!-- View Toggle -->
        <div class="flex p-0.5 bg-dark-800 border border-dark-700 rounded-xl mr-2">
          <button 
            type="button"
            class="px-3 py-1 text-xs font-semibold rounded-lg transition {activeTab === 'editor' ? 'bg-brand-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'}"
            on:click={() => activeTab = 'editor'}
          >
            Code Editor
          </button>
          <button 
            type="button"
            class="px-3 py-1 text-xs font-semibold rounded-lg transition {activeTab === 'inspector' ? 'bg-brand-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'}"
            on:click={() => activeTab = 'inspector'}
          >
            Structure Inspector
          </button>
        </div>

        <button
          type="button"
          class="p-2 rounded-xl bg-dark-800 hover:bg-dark-700 border border-dark-700 text-slate-300 hover:text-white transition"
          title="Copy Content"
          on:click={copyContent}
        >
          <Copy class="w-4 h-4" />
        </button>

        <button
          type="button"
          disabled={!isDirty || isSaving}
          class="flex items-center gap-1.5 px-4 py-2 rounded-xl bg-brand-600 hover:bg-brand-500 disabled:opacity-40 text-white text-xs font-bold transition shadow-lg shadow-brand-600/20"
          on:click={handleSave}
        >
          <Save class="w-4 h-4 {isSaving ? 'animate-spin' : ''}" />
          <span>{isSaving ? 'Saving...' : 'Save (Ctrl+S)'}</span>
        </button>

        <button 
          type="button"
          class="p-2 rounded-xl bg-dark-800 hover:bg-rose-500/20 text-slate-400 hover:text-rose-300 border border-dark-700 transition"
          on:click={() => {
            if (!isDirty || confirm('Discard unsaved changes?')) {
              onClose();
            }
          }}
        >
          <X class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 flex overflow-hidden">
      {#if isLoading}
        <div class="flex-1 flex flex-col items-center justify-center gap-3">
          <RefreshCw class="w-8 h-8 text-brand-400 animate-spin" />
          <p class="text-sm text-slate-400">Inspecting and loading file...</p>
        </div>
      {:else if activeTab === 'editor'}
        <!-- Section Navigator Sidebar -->
        {#if inspection && inspection.sections.length > 0}
          <div class="w-64 border-r border-dark-700/80 bg-dark-950/40 flex flex-col overflow-hidden">
            <div class="p-3 border-b border-dark-800 flex items-center justify-between">
              <span class="text-[11px] font-bold uppercase tracking-wider text-slate-400">Sections ({inspection.sections.length})</span>
            </div>
            <div class="flex-1 overflow-y-auto p-2 space-y-1">
              {#each inspection.sections as sec}
                <button
                  type="button"
                  class="w-full text-left px-2.5 py-1.5 rounded-lg text-xs font-mono transition truncate {selectedSection === sec.name ? 'bg-brand-600 text-white font-bold' : 'text-slate-300 hover:bg-dark-800 hover:text-white'}"
                  on:click={() => jumpToSection(sec.name)}
                >
                  [{sec.name}]
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Textarea Editor with Line Numbers -->
        <div class="flex-1 flex flex-col bg-dark-900 overflow-hidden relative">
          <textarea
            id="code-editor-textarea"
            bind:value={rawText}
            spellcheck="false"
            class="flex-1 w-full p-4 bg-transparent text-slate-200 font-mono text-xs leading-relaxed resize-none focus:outline-none focus:ring-0 selection:bg-brand-600 selection:text-white border-0"
            placeholder="Edit file contents..."
          ></textarea>
        </div>
      {:else if activeTab === 'inspector' && inspection}
        <!-- Structure & Metrics Inspector View -->
        <div class="flex-1 overflow-y-auto p-6 space-y-6">
          
          <!-- Key Metrics Grid -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div class="p-4 rounded-xl bg-dark-800 border border-dark-700/80">
              <span class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Total Sections</span>
              <div class="text-xl font-bold text-white mt-1">{inspection.sections.length}</div>
            </div>
            <div class="p-4 rounded-xl bg-dark-800 border border-dark-700/80">
              <span class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Total Lines</span>
              <div class="text-xl font-bold text-white mt-1">{inspection.total_lines}</div>
            </div>
            <div class="p-4 rounded-xl bg-dark-800 border border-dark-700/80">
              <span class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Category</span>
              <div class="text-xl font-bold text-brand-300 capitalize mt-1">{inspection.category}</div>
            </div>
            <div class="p-4 rounded-xl bg-dark-800 border border-dark-700/80">
              <span class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Key Parameters</span>
              <div class="text-xl font-bold text-indigo-300 mt-1">{Object.keys(inspection.key_values).length}</div>
            </div>
          </div>

          <!-- Animation Actions Inspector -->
          {#if inspection.anim_actions && inspection.anim_actions.length > 0}
            <div class="space-y-3">
              <h3 class="text-xs font-bold uppercase tracking-wider text-slate-300 flex items-center gap-2">
                <Sparkles class="w-4 h-4 text-purple-400" />
                <span>Animation Actions ({inspection.anim_actions.length})</span>
              </h3>
              <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                {#each inspection.anim_actions as act}
                  <div class="p-3 rounded-xl bg-dark-800/80 border border-dark-700/80 space-y-1">
                    <div class="flex items-center justify-between">
                      <span class="font-mono font-bold text-xs text-purple-300">Action {act.action_no}</span>
                      <span class="text-[10px] text-slate-400">{act.frame_count} frames • {act.total_ticks} ticks</span>
                    </div>
                    {#if act.description}
                      <p class="text-xs text-slate-300 truncate">{act.description}</p>
                    {/if}
                    <div class="flex gap-1.5 pt-1">
                      {#if act.has_hitbox}
                        <span class="px-1.5 py-0.5 rounded text-[9px] font-bold bg-rose-500/20 text-rose-300 border border-rose-500/30">Hitbox</span>
                      {/if}
                      {#if act.has_hurtbox}
                        <span class="px-1.5 py-0.5 rounded text-[9px] font-bold bg-blue-500/20 text-blue-300 border border-blue-500/30">Hurtbox</span>
                      {/if}
                      {#if act.has_loop}
                        <span class="px-1.5 py-0.5 rounded text-[9px] font-bold bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">Loop</span>
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Commands Inspector -->
          {#if inspection.commands && inspection.commands.length > 0}
            <div class="space-y-3">
              <h3 class="text-xs font-bold uppercase tracking-wider text-slate-300 flex items-center gap-2">
                <Zap class="w-4 h-4 text-cyan-400" />
                <span>Registered Commands ({inspection.commands.length})</span>
              </h3>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                {#each inspection.commands as cmd}
                  <div class="p-3 rounded-xl bg-dark-800/80 border border-dark-700/80 flex items-center justify-between">
                    <div>
                      <div class="font-mono font-bold text-xs text-cyan-300">{cmd.name}</div>
                      <div class="text-[11px] font-mono text-slate-400">{cmd.command}</div>
                    </div>
                    <div class="text-right text-[10px] text-slate-400 font-mono">
                      <div>Time: {cmd.time}</div>
                      <div>Buffer: {cmd.buffer_time}</div>
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          <!-- StateDef Inspector -->
          {#if inspection.state_defs && inspection.state_defs.length > 0}
            <div class="space-y-3">
              <h3 class="text-xs font-bold uppercase tracking-wider text-slate-300 flex items-center gap-2">
                <Cpu class="w-4 h-4 text-amber-400" />
                <span>State Definitions ({inspection.state_defs.length})</span>
              </h3>
              <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
                {#each inspection.state_defs as st}
                  <div class="p-3 rounded-xl bg-dark-800/80 border border-dark-700/80 flex items-center justify-between">
                    <div>
                      <div class="font-mono font-bold text-xs text-amber-300">StateDef {st.state_no}</div>
                      {#if st.name}
                        <div class="text-[11px] text-slate-300 truncate max-w-[140px]">{st.name}</div>
                      {/if}
                    </div>
                    <div class="text-right text-[10px] text-slate-400 font-mono">
                      <div>Type: {st.type}/{st.move_type}/{st.physics}</div>
                      <div>Controllers: {st.controller_count}</div>
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Key Values Table -->
          {#if Object.keys(inspection.key_values).length > 0}
            <div class="space-y-3">
              <h3 class="text-xs font-bold uppercase tracking-wider text-slate-300">Key Parameters Table</h3>
              <div class="rounded-xl border border-dark-700/80 overflow-hidden">
                <table class="w-full text-left text-xs">
                  <thead class="bg-dark-950/60 border-b border-dark-800 text-slate-400 text-[10px] font-bold uppercase">
                    <tr>
                      <th class="py-2.5 px-4">Parameter</th>
                      <th class="py-2.5 px-4">Value</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-dark-800">
                    {#each Object.entries(inspection.key_values) as [k, v]}
                      <tr class="hover:bg-dark-800/50">
                        <td class="py-2 px-4 font-mono font-bold text-slate-300">{k}</td>
                        <td class="py-2 px-4 font-mono text-slate-400 truncate max-w-md">{v}</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </div>
          {/if}

        </div>
      {/if}
    </div>

  </div>
</div>
