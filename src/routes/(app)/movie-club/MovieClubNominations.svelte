<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import type { MovieClubCycleResponse, Content } from "@/types";
	import { nominateMovie, removeNomination } from "@/lib/util/api";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import Modal from "@/lib/Modal.svelte";
	import SearchMovieModal from "./SearchMovieModal.svelte";

	export let cycleData: MovieClubCycleResponse;

	const dispatch = createEventDispatcher<{ nominationChanged: void }>();

	let showSearchModal = false;
	let submitting = false;

	async function handleNomination(content: Content, reason: string) {
		if (submitting) return;
		
		// Frontend validation to prevent nominations beyond limit
		if (!canNominateLocal) {
			console.warn("Nomination blocked: limit reached or not allowed");
			return;
		}

		submitting = true;
		const success = await nominateMovie({
			contentId: content.tmdbId,
			reason: reason
		});

		if (success) {
			dispatch("nominationChanged");
			showSearchModal = false;
		}
		submitting = false;
	}

	async function handleRemoveNomination(nominationId: number) {
		if (submitting) return;

		submitting = true;
		const success = await removeNomination(nominationId);

		if (success) {
			dispatch("nominationChanged");
		}
		submitting = false;
	}

	$: userNominations = cycleData.userNominations;
	$: canNominate = cycleData.canNominate;
	// Calculate max nominations based on existing data
	// TODO: This should come from movie club settings API
	$: maxNominations = cycleData.maxNominations || Math.max(userNominations.length + (canNominate ? 1 : 0), 1);
	// Local validation - check if user has reached the limit
	$: canNominateLocal = canNominate && userNominations.length < maxNominations && !submitting;
</script>

<div class="nominations-section">
	<div class="section-header">
		<h3>
			<Icon icon="add" />
			Movie Nominations
		</h3>
		<p>Nominate movies you'd like the group to consider for this cycle.</p>
	</div>

	<!-- Your Nominations -->
	<div class="your-nominations">
		<h4>Your Nominations ({userNominations.length}/{maxNominations})</h4>
		
		{#if userNominations.length === 0}
			<div class="empty-state">
				<Icon icon="film" />
				<p>You haven't nominated any movies yet.</p>
				{#if canNominateLocal}
					<button class="nominate-btn" on:click={() => showSearchModal = true}>
						<Icon icon="add" />
						Nominate a Movie
					</button>
				{:else}
					<p class="limit-reached">Nomination limit reached</p>
				{/if}
			</div>
		{:else}
			<div class="nominations-grid">
				{#each userNominations as nomination}
					<div class="nomination-card">
						<div class="poster-container">
							<Poster media={nomination.content} showRating={false} disableInteraction={true} />
						</div>
						<div class="nomination-details">
							<h5>{nomination.content?.title}</h5>
							{#if nomination.reason}
								<p class="reason">"{nomination.reason}"</p>
							{/if}
							<button 
								class="remove-btn"
								on:click={() => handleRemoveNomination(nomination.id)}
								disabled={submitting}
							>
								<Icon icon="trash" />
								Remove
							</button>
						</div>
					</div>
				{/each}
			</div>

			{#if canNominateLocal}
				<button class="nominate-btn" on:click={() => showSearchModal = true}>
					<Icon icon="add" />
					Nominate Another Movie
				</button>
			{/if}
		{/if}
	</div>

	<!-- All Nominations -->
	{#if cycleData.cycle.nominations && cycleData.cycle.nominations.length > 0}
		<div class="all-nominations">
			<h4>All Nominations ({cycleData.cycle.nominations.length})</h4>
			<div class="nominations-grid">
				{#each cycleData.cycle.nominations as nomination}
					<div class="nomination-card">
						<div class="poster-container">
							<Poster media={nomination.content} showRating={false} disableInteraction={true} />
						</div>
						<div class="nomination-details">
							<h5>{nomination.content?.title}</h5>
							<div class="nominators">
								<p class="nominator-label">
									Nominated by: 
									{#each nomination.nominators as nominator, i}
										<span class="nominator">{nominator.username}</span>{#if i < nomination.nominators.length - 1}, {/if}
									{/each}
								</p>
							</div>
							{#if nomination.reasons && nomination.reasons.length > 0}
								<div class="reasons">
									{#each nomination.reasons as reason, i}
										<p class="reason">"{reason}"</p>
									{/each}
								</div>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

{#if showSearchModal}
	<Modal on:close={() => showSearchModal = false}>
		<SearchMovieModal 
			on:movieSelected={(e) => handleNomination(e.detail.content, e.detail.reason)}
			on:close={() => showSearchModal = false}
		/>
	</Modal>
{/if}

<style lang="scss">
	.nominations-section {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.section-header {
		text-align: center;

		h3 {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 0.5rem;
			margin-bottom: 0.5rem;
			font-size: 1.5rem;
		}

		p {
			color: var(--text-muted);
			margin: 0;
		}
	}

	.your-nominations, .all-nominations {
		h4 {
			margin-bottom: 1rem;
			padding-bottom: 0.5rem;
			border-bottom: 1px solid var(--border);
		}
	}

	.empty-state {
		text-align: center;
		padding: 3rem 1rem;
		color: var(--text-muted);

		:global(svg) {
			font-size: 3rem;
			margin-bottom: 1rem;
			opacity: 0.5;
		}

		p {
			margin-bottom: 1.5rem;
		}

		.limit-reached {
			color: var(--warning);
			font-weight: 500;
		}
	}

	.nominations-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.nomination-card {
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: 8px;
		overflow: hidden;
		transition: transform 0.2s ease, box-shadow 0.2s ease;

		&:hover {
			transform: translateY(-2px);
			box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
		}

		.poster-container {
			aspect-ratio: 2/3;
			overflow: hidden;
		}

		.nomination-details {
			padding: 1rem;

			h5 {
				margin: 0 0 0.5rem 0;
				font-size: 0.9rem;
				font-weight: 600;
				line-height: 1.2;
			}

			.nominators {
				margin: 0 0 0.5rem 0;
			}

			.nominator-label {
				font-size: 0.8rem;
				color: var(--text-muted);
				margin: 0;
				font-weight: 500;
			}

			.nominator {
				color: var(--primary);
				font-weight: 600;
			}

			.reasons {
				margin: 0 0 1rem 0;
			}

			.reason {
				font-size: 0.8rem;
				color: var(--text-muted);
				font-style: italic;
				margin: 0 0 0.5rem 0;
				line-height: 1.3;
			}

			.reason:last-child {
				margin-bottom: 1rem;
			}

			.remove-btn {
				background: var(--danger);
				color: white;
				border: none;
				padding: 0.5rem 0.75rem;
				border-radius: 4px;
				font-size: 0.8rem;
				cursor: pointer;
				display: flex;
				align-items: center;
				gap: 0.25rem;
				width: 100%;
				justify-content: center;

				&:hover:not(:disabled) {
					background: var(--danger-dark);
				}

				&:disabled {
					opacity: 0.6;
					cursor: not-allowed;
				}
			}
		}
	}

	.nominate-btn {
		background: var(--primary);
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 6px;
		font-size: 1rem;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin: 0 auto;

		&:hover {
			background: var(--primary-dark);
		}
	}

	@media (max-width: 768px) {
		.nominations-grid {
			grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
			gap: 0.75rem;
		}

		.nomination-card .nomination-details {
			padding: 0.75rem;

			h5 {
				font-size: 0.85rem;
			}

			.reason, .nominator {
				font-size: 0.75rem;
			}
		}
	}
</style>