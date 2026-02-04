<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$lib/shared/utils/helpers';

  export let method: 'get' | 'post' | 'put' | 'patch' | 'delete' = 'post';
  export let action = '';
  export let loading = false;
  export let disabled = false;
  export let validateOnSubmit = true;
  export let validateOnBlur = false;
  export let validateOnChange = false;
  export let resetOnSuccess = false;

  const dispatch = createEventDispatcher<{
    submit: { formData: FormData; data: Record<string, unknown> };
    success: { data: Record<string, unknown> };
    error: { errors: Record<string, string>; message: string };
    reset: void;
  }>();

  let formElement: HTMLFormElement;
  let errors: Record<string, string> = {};
  let touched: Record<string, boolean> = {};

  export function reset() {
    formElement?.reset();
    errors = {};
    touched = {};
    dispatch('reset');
  }

  export function setError(field: string, message: string) {
    errors = { ...errors, [field]: message };
  }

  export function setErrors(newErrors: Record<string, string>) {
    errors = { ...errors, ...newErrors };
  }

  export function clearError(field: string) {
    const { [field]: _, ...rest } = errors;
    errors = rest;
  }

  export function clearErrors() {
    errors = {};
  }

  export function getValues(): Record<string, unknown> {
    if (!formElement) return {};
    const formData = new FormData(formElement);
    const data: Record<string, unknown> = {};
    
    formData.forEach((value, key) => {
      if (data[key] !== undefined) {
        if (Array.isArray(data[key])) {
          (data[key] as unknown[]).push(value);
        } else {
          data[key] = [data[key], value];
        }
      } else {
        data[key] = value;
      }
    });
    
    return data;
  }

  function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    
    if (!formElement) return;

    const formData = new FormData(formElement);
    const data = getValues();

    // Mark all fields as touched on submit
    const fields = Array.from(formElement.elements)
      .filter((el): el is HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement => 
        el instanceof HTMLInputElement || 
        el instanceof HTMLSelectElement || 
        el instanceof HTMLTextAreaElement
      )
      .map(el => el.name)
      .filter(Boolean);
    
    touched = fields.reduce((acc, field) => ({ ...acc, [field]: true }), {});

    dispatch('submit', { formData, data });
  }

  function handleFieldBlur(event: FocusEvent) {
    if (!validateOnBlur) return;
    
    const target = event.target as HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement;
    if (target.name) {
      touched = { ...touched, [target.name]: true };
    }
  }

  function handleFieldChange(event: Event) {
    if (!validateOnChange) return;
    
    const target = event.target as HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement;
    if (target.name && touched[target.name]) {
      // Clear error when field is modified
      clearError(target.name);
    }
  }

  $: isSubmitDisabled = disabled || loading || (validateOnSubmit && Object.keys(errors).length > 0);
</script>

<form
  bind:this={formElement}
  {method}
  {action}
  class={cn('space-y-6', $$props.class)}
  on:submit={handleSubmit}
  on:blur|capture={handleFieldBlur}
  on:input|capture={handleFieldChange}
  novalidate
>
  <slot 
    {errors} 
    {touched} 
    {loading} 
    {getValues}
    {setError}
    {clearError}
    {reset}
  />
  
  {#if $$slots.footer}
    <div class="flex items-center justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
      <slot name="footer" {loading} disabled={isSubmitDisabled} {reset} />
    </div>
  {/if}
</form>
