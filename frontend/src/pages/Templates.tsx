import { useState, useEffect } from 'react'
import { Plus, Trash2, Edit2, Eye } from 'lucide-react'
import Card from '@/components/common/Card'
import Button from '@/components/common/Button'
import Input from '@/components/common/Input'
import { templateApi } from '@/services/api'

interface Template {
  id: number
  name: string
  description: string
  content: string
  variables: string[]
  created_at: string
  updated_at: string
}

interface TemplatePreview {
  phone: string
  message: string
}

export default function Templates() {
  const [templates, setTemplates] = useState<Template[]>([])
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [showPreview, setShowPreview] = useState(false)
  const [previewData, setPreviewData] = useState<TemplatePreview[]>([])

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    content: '',
  })

  const [previewRecipients, setPreviewRecipients] = useState('')

  useEffect(() => {
    loadTemplates()
  }, [])

  const loadTemplates = async () => {
    try {
      setLoading(true)
      const response = await templateApi.getAll()
      setTemplates(response.data || [])
    } catch (error) {
      console.error('Failed to load templates:', error)
      setTemplates([])
      showMessage('error', 'Failed to load templates')
    } finally {
      setLoading(false)
    }
  }

  const showMessage = (type: 'success' | 'error', text: string) => {
    setMessage({ type, text })
    setTimeout(() => setMessage(null), 3000)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!formData.name.trim() || !formData.content.trim()) {
      showMessage('error', 'Name and content are required')
      return
    }

    try {
      setLoading(true)
      if (editingId) {
        await templateApi.update(editingId, formData)
        showMessage('success', 'Template updated')
      } else {
        await templateApi.create(formData)
        showMessage('success', 'Template created')
      }
      resetForm()
      loadTemplates()
    } catch (error: any) {
      showMessage('error', error.message || 'Failed to save template')
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = (template: Template) => {
    setFormData({
      name: template.name,
      description: template.description,
      content: template.content,
    })
    setEditingId(template.id)
    setShowForm(true)
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this template?')) return

    try {
      setLoading(true)
      await templateApi.delete(id)
      showMessage('success', 'Template deleted')
      loadTemplates()
    } catch (error: any) {
      showMessage('error', error.message || 'Failed to delete template')
    } finally {
      setLoading(false)
    }
  }

  const handlePreview = async (template: Template) => {
    const phones = previewRecipients
      .split('\n')
      .map(p => p.trim())
      .filter(p => p)

    if (phones.length === 0) {
      showMessage('error', 'Enter at least one phone number')
      return
    }

    try {
      setLoading(true)
      const recipients = phones.map(phone => ({
        phone,
        variables: {},
      }))
      const response = await templateApi.preview(template.id, recipients)
      setPreviewData(response.data)
      setShowPreview(true)
    } catch (error: any) {
      showMessage('error', error.message || 'Failed to preview template')
    } finally {
      setLoading(false)
    }
  }

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      content: '',
    })
    setEditingId(null)
    setShowForm(false)
  }

  const extractVariables = (content: string) => {
    const regex = /\{\{([^}]+)\}\}/g
    const matches = []
    let match
    while ((match = regex.exec(content)) !== null) {
      matches.push(match[1])
    }
    return [...new Set(matches)]
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Message Templates</h1>
          <p className="text-gray-600 mt-1">Create reusable templates with variables to personalize messages</p>
        </div>
        <Button onClick={() => setShowForm(!showForm)}>
          <Plus className="w-4 h-4 mr-2" />
          New Template
        </Button>
      </div>

      {message && (
        <div
          className={`rounded-lg p-4 ${
            message.type === 'success'
              ? 'bg-green-50 border border-green-200 text-green-800'
              : 'bg-red-50 border border-red-200 text-red-800'
          }`}
        >
          {message.text}
        </div>
      )}

      {showForm && (
        <Card>
          <div className="p-6">
            <h2 className="text-lg font-semibold mb-4">
              {editingId ? 'Edit Template' : 'Create Template'}
            </h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <Input
                label="Template Name"
                placeholder="e.g., Welcome Message"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                disabled={!!editingId}
              />

              <Input
                label="Description"
                placeholder="What is this template for?"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              />

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Message Content
                </label>
                <textarea
                  value={formData.content}
                  onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                  placeholder="Use {{variable}} for personalization. Example: Hello {{name}}, welcome to {{company}}!"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  rows={6}
                />
              </div>

              {formData.content && (
                <div className="bg-blue-50 border border-blue-200 rounded p-3">
                  <p className="text-sm font-semibold text-blue-900 mb-2">Variables found:</p>
                  <div className="flex flex-wrap gap-2">
                    {extractVariables(formData.content).map((variable) => (
                      <span key={variable} className="bg-blue-200 text-blue-800 px-2 py-1 rounded text-xs font-mono">
                        {variable}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              <div className="flex space-x-3">
                <Button type="submit" isLoading={loading}>
                  {editingId ? 'Update' : 'Create'} Template
                </Button>
                <Button variant="secondary" onClick={resetForm}>
                  Cancel
                </Button>
              </div>
            </form>
          </div>
        </Card>
      )}

      {showPreview && (
        <Card>
          <div className="p-6">
            <h2 className="text-lg font-semibold mb-4">Template Preview</h2>
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {previewData.map((preview, idx) => (
                <div key={idx} className="bg-gray-50 p-3 rounded border border-gray-200">
                  <p className="text-sm font-mono text-gray-600 mb-1">{preview.phone}</p>
                  <p className="text-sm text-gray-800">{preview.message}</p>
                </div>
              ))}
            </div>
            <Button variant="secondary" onClick={() => setShowPreview(false)} className="mt-4">
              Close Preview
            </Button>
          </div>
        </Card>
      )}

      <div className="grid gap-4">
        {loading && templates.length === 0 ? (
          <Card>
            <div className="p-6 text-center text-gray-500">Loading templates...</div>
          </Card>
        ) : templates.length === 0 ? (
          <Card>
            <div className="p-6 text-center text-gray-500">No templates yet</div>
          </Card>
        ) : (
          templates.map((template) => (
            <Card key={template.id}>
              <div className="p-6">
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <h3 className="font-semibold text-lg">{template.name}</h3>
                    {template.description && (
                      <p className="text-sm text-gray-600">{template.description}</p>
                    )}
                  </div>
                </div>

                <div className="bg-gray-50 p-3 rounded mb-4 text-sm">
                  <p className="text-gray-700 whitespace-pre-wrap">{template.content}</p>
                </div>

                {template.variables.length > 0 && (
                  <div className="mb-4">
                    <p className="text-xs font-semibold text-gray-600 mb-2">Variables:</p>
                    <div className="flex flex-wrap gap-2">
                      {template.variables.map((variable) => (
                        <span key={variable} className="bg-blue-100 text-blue-800 px-2 py-1 rounded text-xs font-mono">
                          {variable}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                <div className="flex space-x-2">
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => {
                      setPreviewRecipients('')
                      handlePreview(template)
                    }}
                  >
                    <Eye className="w-4 h-4" />
                  </Button>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => handleEdit(template)}
                  >
                    <Edit2 className="w-4 h-4" />
                  </Button>
                  <Button
                    variant="danger"
                    size="sm"
                    onClick={() => handleDelete(template.id)}
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </Card>
          ))
        )}
      </div>
    </div>
  )
}
