<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import type { MovieClubCycleResponse } from "@/types";
	import MovieClubPhaseStatus from "./MovieClubPhaseStatus.svelte";
	import MovieClubNominations from "./MovieClubNominations.svelte";
	import MovieClubVoting from "./MovieClubVoting.svelte";
	import MovieClubResults from "./MovieClubResults.svelte";

	export let cycleData: MovieClubCycleResponse;

	const dispatch = createEventDispatcher<{
		refresh: void;
		openSearchModal: void;
		nominateMovie: { content: any; reason: string };
	}>();

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
				on:openSearchModal={() => dispatch("openSearchModal")}
				on:nominateMovie={(e) => dispatch("nominateMovie", e.detail)}
			/>
		{:else if currentPhase === "voting"}
			<MovieClubVoting {cycleData} on:votesChanged={onDataChanged} />
		{:else if currentPhase === "watching"}
			<MovieClubResults {cycleData} />
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
		--shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1),
			0 2px 4px -1px rgba(0, 0, 0, 0.06);
		--shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1),
			0 4px 6px -2px rgba(0, 0, 0, 0.05);

		@media (prefers-color-scheme: dark) {
			--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.3);
			--shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.4),
				0 2px 4px -1px rgba(0, 0, 0, 0.3);
			--shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.5),
				0 4px 6px -2px rgba(0, 0, 0, 0.4);
		}
	}

	.phase-content {
		background: var(--background);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		border: 1px solid var(--border);
		transition: all 0.2s ease;
		position: relative;

		&::before {
			content: "";
			position: absolute;
			inset: 0;
			background: var(--primary);
			opacity: 0.01;
			border-radius: inherit;
			pointer-events: none;
		}

		&:hover {
			border-color: var(--primary);
			box-shadow: var(--shadow-sm);

			&::before {
				opacity: 0.02;
			}
		}
	}

	@media (max-width: 768px) {
		.phase-content {
			padding: var(--space-md);
		}
	}
</style>
