<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import type { MovieClubCycleResponse, MovieClubVoteItem } from "@/types";
	import { voteForMovies, clearVotes } from "@/lib/util/api";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";

	// Component for collapsible reason text
	function CollapsibleReason(reason: string, maxLength: number = 40) {
		return {
			shouldCollapse: reason.length > maxLength,
			truncated: reason.substring(0, maxLength) + "...",
			full: reason
		};
	}

	export let cycleData: MovieClubCycleResponse;

	const dispatch = createEventDispatcher<{ votesChanged: void }>();

	let submitting = false;
	let selectedVotes: MovieClubVoteItem[] = [];
	let expandedReasons: Set<string> = new Set(); // Track expanded reasons by unique ID

	// Initialize selected votes from current user votes
	$: {
		selectedVotes = cycleData.userVotes.map(vote => ({
			contentId: vote.contentId,
			priority: vote.priority
		}));
	}

	$: nominations = cycleData.cycle.nominations || [];
	$: maxVotes = cycleData.maxVotes || 2; // Get from settings, fallback to 2
	$: canVote = cycleData.canVote;

	function toggleVote(contentId: number) {
		const existingIndex = selectedVotes.findIndex(v => v.contentId === contentId);
		
		if (existingIndex >= 0) {
			// Remove vote
			const removedPriority = selectedVotes[existingIndex].priority;
			selectedVotes = selectedVotes
				.filter(v => v.contentId !== contentId)
				.map(v => v.priority > removedPriority ? { ...v, priority: v.priority - 1 } : v);
		} else if (selectedVotes.length < maxVotes) {
			// Add vote with next priority
			const nextPriority = selectedVotes.length + 1;
			selectedVotes = [...selectedVotes, { contentId, priority: nextPriority }];
		}
	}

	function getVotePriority(contentId: number): number | null {
		const vote = selectedVotes.find(v => v.contentId === contentId);
		return vote ? vote.priority : null;
	}

	function moveVotePriority(contentId: number, direction: 'up' | 'down') {
		const currentVote = selectedVotes.find(v => v.contentId === contentId);
		if (!currentVote) return;

		const newPriority = direction === 'up' ? currentVote.priority - 1 : currentVote.priority + 1;
		
		if (newPriority < 1 || newPriority > selectedVotes.length) return;

		// Swap priorities
		const otherVote = selectedVotes.find(v => v.priority === newPriority);
		if (otherVote) {
			selectedVotes = selectedVotes.map(v => {
				if (v.contentId === contentId) return { ...v, priority: newPriority };
				if (v.contentId === otherVote.contentId) return { ...v, priority: currentVote.priority };
				return v;
			});
		}
	}

	async function submitVotes() {
		if (submitting || selectedVotes.length === 0) return;

		submitting = true;
		const success = await voteForMovies({ votes: selectedVotes, cycleId: cycleData.cycle.id });
		
		if (success) {
			dispatch("votesChanged");
		}
		submitting = false;
	}

	async function handleClearVotes() {
		if (submitting) return;

		submitting = true;
		const success = await clearVotes(cycleData.cycle.id);
		
		if (success) {
			selectedVotes = [];
			dispatch("votesChanged");
		}
		submitting = false;
	}

	function getPriorityLabel(priority: number): string {
		switch (priority) {
			case 1: return "1st Choice";
			case 2: return "2nd Choice";
			case 3: return "3rd Choice";
			default: return `${priority}th Choice`;
		}
	}

	function toggleReasonExpansion(reasonId: string) {
		if (expandedReasons.has(reasonId)) {
			expandedReasons.delete(reasonId);
		} else {
			expandedReasons.add(reasonId);
		}
		expandedReasons = expandedReasons; // Trigger reactivity
	}
</script>

<div class="voting-section">
	<div class="section-header">
		<h3>
			<Icon icon="check" />
			Vote for Movies
		</h3>
		<p>Choose your top {maxVotes} movies from the nominations. Order matters!</p>
	</div>

	{#if nominations.length === 0}
		<div class="no-nominations">
			<Icon icon="film" />
			<p>No movies have been nominated yet.</p>
		</div>
	{:else}
		<div class="voting-controls">
			<div class="vote-status">
				<span class="votes-count">
					{selectedVotes.length}/{maxVotes} votes cast
				</span>
				{#if selectedVotes.length > 0}
					<button 
						class="clear-votes-btn"
						on:click={handleClearVotes}
						disabled={submitting}
					>
						<Icon icon="trash" />
						Clear Votes
					</button>
				{/if}
			</div>

			{#if selectedVotes.length > 0}
				<button 
					class="submit-votes-btn"
					on:click={submitVotes}
					disabled={submitting}
				>
					<Icon icon="check" />
					{submitting ? "Submitting..." : "Submit Votes"}
				</button>
			{/if}
		</div>

		<div class="nominations-grid">
			{#each nominations as nomination}
				{@const priority = getVotePriority(nomination.contentId)}
				{@const isSelected = priority !== null}
				
				<div 
					class="nomination-card"
					class:selected={isSelected}
					on:click={() => toggleVote(nomination.contentId)}
				>
					{#if isSelected}
						<div class="priority-badge">
							{getPriorityLabel(priority)}
						</div>
					{/if}

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
									{@const reasonData = CollapsibleReason(reason)}
									{@const reasonId = `vote-${nomination.contentId}-${i}`}
									{@const isExpanded = expandedReasons.has(reasonId)}
									
									<div class="reason-container">
										<p class="reason">
											"{isExpanded ? reasonData.full : (reasonData.shouldCollapse ? reasonData.truncated : reasonData.full)}"
										</p>
										{#if reasonData.shouldCollapse}
											<button 
												class="expand-btn" 
												on:click|stopPropagation={() => toggleReasonExpansion(reasonId)}
												type="button"
											>
												{isExpanded ? 'Show less' : 'Show more'}
											</button>
										{/if}
									</div>
								{/each}
							</div>
						{/if}

						{#if isSelected && selectedVotes.length > 1}
							<div class="priority-controls">
								<button 
									class="priority-btn"
									class:disabled={priority === 1}
									on:click|stopPropagation={() => moveVotePriority(nomination.contentId, 'up')}
									disabled={priority === 1}
								>
									<Icon icon="chevron" />
									Higher
								</button>
								<button 
									class="priority-btn"
									class:disabled={priority === selectedVotes.length}
									on:click|stopPropagation={() => moveVotePriority(nomination.contentId, 'down')}
									disabled={priority === selectedVotes.length}
								>
									<Icon icon="chevron" />
									Lower
								</button>
							</div>
						{/if}
					</div>
				</div>
			{/each}
		</div>

		{#if selectedVotes.length > 0}
			<div class="vote-summary">
				<h4>Your Vote Summary</h4>
				<div class="vote-list">
					{#each selectedVotes.sort((a, b) => a.priority - b.priority) as vote}
						{@const nomination = nominations.find(n => n.contentId === vote.contentId)}
						<div class="vote-item">
							<span class="priority">{getPriorityLabel(vote.priority)}</span>
							<span class="title">{nomination?.content?.title}</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>

<style lang="scss">
	.voting-section {
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

	.no-nominations {
		text-align: center;
		padding: 3rem 1rem;
		color: var(--text-muted);

		:global(svg) {
			font-size: 3rem;
			margin-bottom: 1rem;
			opacity: 0.5;
		}
	}

	.voting-controls {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 1rem;
		padding: 1rem;
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: 8px;
	}

	.vote-status {
		display: flex;
		align-items: center;
		gap: 1rem;

		.votes-count {
			font-weight: 600;
			color: var(--primary);
		}

		.clear-votes-btn {
			background: var(--danger);
			color: white;
			border: none;
			padding: 0.5rem 1rem;
			border-radius: 4px;
			font-size: 0.875rem;
			cursor: pointer;
			display: flex;
			align-items: center;
			gap: 0.25rem;

			&:hover:not(:disabled) {
				background: var(--danger-dark);
			}

			&:disabled {
				opacity: 0.6;
				cursor: not-allowed;
			}
		}
	}

	.submit-votes-btn {
		background: var(--success);
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 6px;
		font-size: 1rem;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;

		&:hover:not(:disabled) {
			background: var(--success-dark);
		}

		&:disabled {
			opacity: 0.6;
			cursor: not-allowed;
		}
	}

	.nominations-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 1rem;
	}

	.nomination-card {
		background: var(--background);
		border: 2px solid var(--border);
		border-radius: 8px;
		overflow: hidden;
		cursor: pointer;
		transition: all 0.2s ease;
		position: relative;

		&:hover {
			transform: translateY(-2px);
			box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
		}

		&.selected {
			border-color: var(--primary);
			box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
		}

		.priority-badge {
			position: absolute;
			top: 0.5rem;
			right: 0.5rem;
			background: var(--primary);
			color: white;
			padding: 0.25rem 0.5rem;
			border-radius: 4px;
			font-size: 0.75rem;
			font-weight: 600;
			z-index: 2;
		}

		.poster-container {
			position: relative;
			aspect-ratio: 2/3;
			overflow: hidden;
			display: flex;
			align-items: center;
			justify-content: center;
			
			:global(li) {
				list-style: none;
				margin: 0;
				padding: 0;
				width: 100%;
				height: 100%;
				display: flex;
				align-items: center;
				justify-content: center;
			}
			
			:global(.container) {
				width: 100% !important;
				height: 100% !important;
				display: flex !important;
				align-items: center !important;
				justify-content: center !important;
			}
			
			:global(img) {
				max-width: 100%;
				max-height: 100%;
				object-fit: cover;
				object-position: center;
			}
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

			.reason-container {
				margin-bottom: 0.5rem;
			}

			.reason-container:last-child {
				margin-bottom: 1rem;
			}

			.reason {
				font-size: 0.8rem;
				color: var(--text-muted);
				font-style: italic;
				margin: 0 0 0.25rem 0;
				line-height: 1.3;
			}

			.expand-btn {
				background: none;
				border: none;
				color: var(--primary);
				font-size: 0.7rem;
				cursor: pointer;
				padding: 0;
				text-decoration: underline;
				font-family: inherit;
				
				&:hover {
					color: var(--primary-dark);
				}
			}

			.priority-controls {
				display: flex;
				gap: 0.5rem;
				margin-top: 0.75rem;

				.priority-btn {
					flex: 1;
					background: var(--background-secondary);
					border: 1px solid var(--border);
					padding: 0.5rem;
					border-radius: 4px;
					font-size: 0.75rem;
					cursor: pointer;
					display: flex;
					align-items: center;
					justify-content: center;
					gap: 0.25rem;

					&:first-child :global(svg) {
						transform: rotate(180deg);
					}

					&:last-child :global(svg) {
						transform: rotate(0deg);
					}

					&:hover:not(:disabled) {
						background: var(--background);
					}

					&:disabled, &.disabled {
						opacity: 0.5;
						cursor: not-allowed;
					}
				}
			}
		}
	}

	.vote-summary {
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 1.5rem;

		h4 {
			margin: 0 0 1rem 0;
			color: var(--primary);
		}

		.vote-list {
			display: flex;
			flex-direction: column;
			gap: 0.5rem;
		}

		.vote-item {
			display: flex;
			align-items: center;
			gap: 1rem;

			.priority {
				font-weight: 600;
				color: var(--primary);
				min-width: 80px;
			}

			.title {
				color: var(--text);
			}
		}
	}

	@media (max-width: 768px) {
		.voting-controls {
			flex-direction: column;
			align-items: stretch;
		}

		.vote-status {
			justify-content: space-between;
		}

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