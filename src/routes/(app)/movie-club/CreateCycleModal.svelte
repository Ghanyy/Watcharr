<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import { onMount } from "svelte";
	import Icon from "@/lib/Icon.svelte";
	import { notify } from "@/lib/util/notify";
	import axios from "axios";

	const dispatch = createEventDispatcher<{
		cycleCreated: void;
		close: void;
	}>();

	let name = "Movie Club Cycle";
	let description = "Choose movies together!";
	let submitting = false;

	onMount(() => {
		// Add escape key listener for closing modal
		const handleKeydown = (e: KeyboardEvent) => {
			if (e.key === "Escape" && !submitting) {
				dispatch("close");
			}
		};

		document.addEventListener("keydown", handleKeydown);

		// Cleanup function
		return () => {
			document.removeEventListener("keydown", handleKeydown);
		};
	});

	async function createCycle() {
		if (submitting || !name.trim()) return;

		submitting = true;
		const nid = notify({ text: "Creating cycle...", type: "loading" });

		try {
			const response = await axios.post("/movie-club/cycle", {
				name: name.trim(),
				description: description.trim(),
			});

			notify({ id: nid, text: "Cycle created successfully!", type: "success" });

			// Small delay to ensure database consistency
			setTimeout(() => {
				dispatch("cycleCreated");
			}, 100);
		} catch (err: any) {
			console.error("Failed to create cycle:", err);
			const message = err.response?.data?.error || "Failed to create cycle";
			notify({ id: nid, text: message, type: "error" });
		} finally {
			submitting = false;
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
			createCycle();
		}
	}
</script>

<div class="create-cycle-modal">
	<div class="modal-header">
		<h3>Create New Movie Club Cycle</h3>
		<button class="close-btn" on:click={() => dispatch("close")}>
			<Icon icon="close" />
		</button>
	</div>

	<div class="modal-content">
		<div class="form-group">
			<label for="cycle-name">Cycle Name</label>
			<input
				id="cycle-name"
				type="text"
				bind:value={name}
				placeholder="Movie Club Cycle"
				maxlength="100"
				disabled={submitting}
				on:keydown={handleKeydown}
			/>
		</div>

		<div class="form-group">
			<label for="cycle-description">Description (Optional)</label>
			<textarea
				id="cycle-description"
				bind:value={description}
				placeholder="Describe this cycle..."
				maxlength="500"
				disabled={submitting}
				on:keydown={handleKeydown}
			></textarea>
		</div>

		<div class="cycle-info">
			<h4>Cycle Information</h4>
			<p>
				This cycle will start immediately in the <strong
					>nomination phase</strong
				>. Users will be able to nominate movies, then vote, and finally see the
				results.
			</p>
			<p>
				Phase durations and user limits can be configured in the server
				settings.
			</p>
		</div>

		<div class="action-buttons">
			<button
				class="cancel-btn"
				on:click={() => dispatch("close")}
				disabled={submitting}
			>
				Cancel
			</button>
			<button
				class="create-btn"
				on:click={createCycle}
				disabled={submitting || !name.trim()}
			>
				{submitting ? "Creating..." : "Create Cycle"}
			</button>
		</div>
	</div>
</div>

<style lang="scss">
	.create-cycle-modal {
		width: 100%;
		height: 100%;
		min-height: 400px;
		background: var(--background);
		border-radius: var(--radius-lg);
		overflow: hidden;
		box-shadow: var(--shadow-lg);
		border: 1px solid var(--border);
		display: flex;
		flex-direction: column;

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

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-lg);
		border-bottom: 1px solid var(--border);

		h3 {
			margin: 0;
			color: var(--text);
			font-weight: 700;
			font-size: 1.25rem;
		}

		.close-btn {
			display: inline-flex;
			align-items: center;
			justify-content: center;
			width: 32px;
			height: 32px;
			background: none;
			border: none;
			cursor: pointer;
			color: var(--text-muted);
			border-radius: var(--radius-md);
			transition: all 0.2s ease;
			border: 1px solid transparent;

			&:hover {
				color: var(--text);
				background: var(--background);
				border-color: var(--border);
				box-shadow: var(--shadow-sm);
			}

			&:focus {
				outline: none;
				box-shadow:
					var(--shadow-md),
					0 0 0 2px var(--primary);
			}
		}
	}

	.modal-content {
		padding: var(--space-lg);
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
		flex: 1;
		min-height: 0;
		overflow-y: auto;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);

		label {
			font-weight: 600;
			color: var(--text);
			font-size: 0.95rem;
		}

		input,
		textarea {
			padding: var(--space-md);
			border: 2px solid var(--border);
			border-radius: var(--radius-lg);
			font-family: inherit;
			font-size: 1rem;
			background: var(--background);
			color: var(--text);
			transition: all 0.2s ease;
			box-shadow: var(--shadow-sm);

			&:focus {
				outline: none;
				border-color: var(--primary);
				box-shadow:
					var(--shadow-md),
					0 0 0 3px rgba(59, 130, 246, 0.1);
			}

			&:disabled {
				opacity: 0.6;
				cursor: not-allowed;
			}

			&::placeholder {
				color: var(--text-muted);
				font-weight: 400;
			}

			&:hover:not(:focus):not(:disabled) {
				box-shadow: var(--shadow-md);
			}
		}

		textarea {
			min-height: 100px;
			resize: vertical;
			line-height: 1.5;
		}
	}

	.cycle-info {
		padding: var(--space-md);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		border-left: 3px solid var(--primary);
		position: relative;

		&::before {
			content: "";
			position: absolute;
			inset: 0;
			background: var(--primary);
			opacity: 0.02;
			border-radius: inherit;
			pointer-events: none;
		}

		h4 {
			margin: 0 0 var(--space-sm) 0;
			color: var(--primary);
			font-weight: 600;
			font-size: 1rem;
		}

		p {
			margin: 0 0 var(--space-sm) 0;
			color: var(--text-muted);
			line-height: 1.5;
			font-size: 0.9rem;

			&:last-child {
				margin-bottom: 0;
			}

			strong {
				color: var(--text);
				font-weight: 600;
			}
		}
	}

	.action-buttons {
		display: flex;
		gap: var(--space-md);
		justify-content: flex-end;
		padding: var(--space-md);
		background: var(--background);
		border-top: 1px solid var(--border);
		border-radius: 0 0 var(--radius-lg) var(--radius-lg);
		margin: var(--space-md) -#{var(--space-lg)} -#{var(--space-lg)} -#{var(
				--space-lg
			)};

		button {
			display: inline-flex;
			align-items: center;
			gap: var(--space-sm);
			padding: var(--space-sm) var(--space-lg);
			border-radius: var(--radius-md);
			font-size: 1rem;
			font-weight: 500;
			cursor: pointer;
			font-family: inherit;
			transition: all 0.2s ease;
			box-shadow: var(--shadow-sm);
			min-width: 120px;
			justify-content: center;

			&:disabled {
				opacity: 0.6;
				cursor: not-allowed;
				transform: none;
				box-shadow: var(--shadow-sm);
			}
		}

		.cancel-btn {
			background: var(--background);
			border: 1px solid var(--border);
			color: var(--text-muted);
			font-size: 0.95rem;
			font-weight: 500;

			&:hover:not(:disabled) {
				color: var(--text);
				background: var(--background-secondary, rgba(0, 0, 0, 0.02));
				border-color: var(--text-muted);
				transform: translateY(-1px);
				box-shadow: var(--shadow-md);
			}

			&:focus {
				outline: none;
				box-shadow:
					var(--shadow-md),
					0 0 0 2px var(--primary);
			}
		}

		.create-btn {
			background: var(--primary);
			border: none;
			color: white;

			&:hover:not(:disabled) {
				background: var(--primary-dark, var(--primary));
				transform: translateY(-1px);
				box-shadow: var(--shadow-md);
			}

			&:focus {
				outline: none;
				box-shadow:
					var(--shadow-md),
					0 0 0 3px rgba(59, 130, 246, 0.2);
			}

			&:active:not(:disabled) {
				transform: translateY(0);
				box-shadow: var(--shadow-sm);
			}
		}
	}

	@media (max-width: 768px) {
		.create-cycle-modal {
			min-height: 300px;
		}

		.modal-header {
			padding: var(--space-md);

			h3 {
				font-size: 1.125rem;
			}

			.close-btn {
				width: 28px;
				height: 28px;
			}
		}

		.modal-content {
			padding: var(--space-md);
			gap: var(--space-md);
		}

		.form-group {
			label {
				font-size: 0.9rem;
			}

			input,
			textarea {
				padding: var(--space-sm) var(--space-md);
				font-size: 0.95rem;
			}

			textarea {
				min-height: 80px;
			}
		}

		.cycle-info {
			padding: var(--space-sm);

			h4 {
				font-size: 0.95rem;
			}

			p {
				font-size: 0.85rem;
			}
		}

		.action-buttons {
			flex-direction: column;
			gap: var(--space-sm);
			padding: var(--space-sm);
			margin: var(--space-sm) -#{var(--space-md)} -#{var(--space-md)} -#{var(
					--space-md
				)};

			button {
				width: 100%;
				justify-content: center;
				padding: var(--space-sm) var(--space-md);
				font-size: 0.9rem;
				min-width: unset;
			}
		}
	}
</style>
