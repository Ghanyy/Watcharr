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
		gap: 2rem;
	}

	.phase-content {
		background: var(--background-secondary);
		border-radius: 12px;
		padding: 1.5rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
	}

	@media (max-width: 768px) {
		.phase-content {
			padding: 1rem;
		}
	}
</style>