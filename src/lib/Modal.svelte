<script lang="ts">
	import Icon from "./Icon.svelte";

	interface Props {
		title: string;
		desc?: string | undefined;
		onClose?: (() => void) | undefined;
		maxWidth?: string;
		error?: string | undefined; // TODO This property is new, mimics what we do with other modals by showing an error at top.. we could migrate to use this in other places.
		children?: import("svelte").Snippet;
	}

	let {
		title,
		desc = undefined,
		onClose = undefined,
		maxWidth = "1000px",
		error = undefined,
		children,
	}: Props = $props();
</script>

<div class="backdrop"></div>
<div class="modal">
	<div style="max-width:{maxWidth};">
		{#if typeof onClose !== "undefined"}
			<button class="close" onclick={onClose}><Icon i="close" wh="20" /></button
			>
		{/if}
		<h3 class="norm">{title}</h3>
		{#if desc}
			<h5 class="norm">{desc}</h5>
		{/if}
		{#if error}
			<span class="error">{error}</span>
		{/if}
		{@render children?.()}
	</div>
</div>

<style lang="scss">
	.backdrop {
		position: fixed;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		backdrop-filter: blur(2px) saturate(180%);
		background-color: color-mix(in srgb, black 85%, transparent);
		z-index: 99998;
	}

	.modal {
		display: flex;
		justify-content: center;
		align-items: center;
		width: 100vw;
		height: 100vh;
		top: 0;
		left: 0;
		position: fixed;
		z-index: 99999;
		padding: 20px;
		box-sizing: border-box;

		& > div {
			position: relative;
			display: flex;
			flex-flow: column;
			min-width: 300px;
			width: 100%;
			max-width: min(90vw, 1200px);
			max-height: calc(100vh - 40px);
			background-color: $bg-color;
			border-radius: 10px;
			padding: 15px 20px;
			outline: 2px solid $text-color;
			overflow: auto;
			box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);

			h5 {
				margin-bottom: 15px;
			}

			button.close {
				position: absolute;
				top: 8px;
				right: 8px;
				width: max-content;
				padding: 3px 5px;
				z-index: 1;
			}

			@media screen and (max-width: 680px) {
				max-width: calc(100vw - 20px);
				max-height: calc(100vh - 20px);
			}

			.error {
				position: sticky;
				top: 0;
				display: flex;
				justify-content: center;
				width: 100%;
				padding: 10px;
				background-color: rgb(221, 48, 48);
				text-transform: capitalize;
				color: white;
				margin-bottom: 15px;
			}
		}
	}
</style>
