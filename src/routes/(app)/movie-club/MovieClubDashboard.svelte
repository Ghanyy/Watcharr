<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import type { MovieClubCycleResponse } from "@/types";
	import MovieClubPhaseStatus from "./MovieClubPhaseStatus.svelte";
	import MovieClubNominations from "./MovieClubNominations.svelte";
	import MovieClubVoting from "./MovieClubVoting.svelte";
	import MovieClubResults from "./MovieClubResults.svelte";

	export let cycleData: MovieClubCycleResponse;

	const dispatch = createEventDispatcher<{ refresh: void }>();

	function onDataChanged() {
		dispatch("refresh");
	}

	$: currentPhase = cycleData.cycle.phase;
</script>

<div class="movie-club-dashboard">
	<MovieClubPhaseStatus cycle={cycleData.cycle} />

	<div class="phase-content">
		{#if currentPhase === "nomination"}
			<MovieClubNominations
				{cycleData}
				on:nominationChanged={onDataChanged}
			/>
		{:else if currentPhase === "voting"}
			<MovieClubVoting
				{cycleData}
				on:votesChanged={onDataChanged}
			/>
		{:else if currentPhase === "watching"}
			<MovieClubResults
				{cycleData}
			/>
		{/if}
	</div>
</div>

<style lang="scss">
	.movie-club-dashboard {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
		
		// CSS custom properties for consistent design system
		--space-xs: 0.25rem;
		--space-sm: 0.5rem;
		--space-md: 1rem;
		--space-lg: 1.5rem;
		--space-xl: 2rem;
		--space-2xl: 3rem;
		
		--radius-sm: 4px;
		--radius-md: 8px;
		--radius-lg: 12px;
		--radius-xl: 16px;
		--radius-full: 50%;
		
		--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
		--shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
		--shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
		
		@media (prefers-color-scheme: dark) {
			--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.3);
			--shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.4), 0 2px 4px -1px rgba(0, 0, 0, 0.3);
			--shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.5), 0 4px 6px -2px rgba(0, 0, 0, 0.4);
		}
	}

	.phase-content {
		background: var(--background-secondary);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		box-shadow: var(--shadow-md);
		border: 1px solid var(--border);
		transition: box-shadow 0.2s ease;
		
		&:hover {
			box-shadow: var(--shadow-lg);
		}
	}

	@media (max-width: 768px) {
		.phase-content {
			padding: var(--space-md);
		}
	}
</style>