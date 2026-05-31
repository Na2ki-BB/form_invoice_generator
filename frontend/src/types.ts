export type Product = {
  id: string
  name: string
  description: string
  unitPrice: number
  minQuantity: number
  maxQuantity: number
  quantity: number
}

export type PublicForm = {
  title: string
  description: string
  publicSlug: string
  products: Omit<Product, 'quantity'>[]
}

export type CustomerInfo = {
  customerName: string
  customerKana: string
  postalCode: string
  address: string
  phone: string
  email: string
  note: string
}
