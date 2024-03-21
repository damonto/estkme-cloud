<script setup lang="ts">
import { FormField, FormItem, FormControl, FormMessage } from '@/components/ui/form'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import * as z from 'zod'
import { PlugZap, Loader2 } from 'lucide-vue-next'
import { useToast } from '@/components/ui/toast/use-toast'

const formSchema = toTypedSchema(
  z.object({
    pinCode: z.string().min(6).max(6)
  })
)

const form = useForm({
  validationSchema: formSchema
})
const { toast } = useToast()
toast({
  title: 'Welcome to the Connect View',
  description: 'Please enter your PIN Code to connect to the server'
})
const onSubmit = form.handleSubmit((values) => {
  console.log(values)
})
</script>

<template>
  <div class="container h-[100vh] flex justify-center items-center">
    <div class="w-full">
      <form @submit="onSubmit" class="space-y-4">
        <FormField v-slot="{ componentField }" name="pinCode">
          <FormItem>
            <FormControl>
              <Input
                minlength="6"
                maxlength="6"
                placeholder="Please enter your PIN Code"
                v-bind="componentField"
              ></Input>
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>
        <Button class="w-full" :disabled="form.isSubmitting.value">
          <template v-if="form.isSubmitting.value">
            <Loader2 class="w-4 h-4 mr-2 animate-spin" />
            Connecting
          </template>
          <template v-else>
            <PlugZap class="w-4 h-4 mr-2" />
            Connect
          </template>
        </Button>
      </form>
    </div>
  </div>
</template>
