'use client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

export default function AdminDashboard() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-blue-50">
        <div className="container mx-auto p-6">
            <div className="text-center mb-8">
            <h1 className="text-4xl font-bold text-gray-900 mb-2">Admin Dashboard</h1>
            </div>
        

        <Card className="max-w-7xl mx-auto shadow-xl border-0">

            
        
        <CardContent>

        {/* Table van bedrijven */}
          <Table>
            <TableHeader>
                {/* Table headers */}
              <TableRow>
                <TableHead>Bedrijfsnaam</TableHead>
                <TableHead>API Key</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>September 2025</TableHead>
                <TableHead>Oktober 2025</TableHead>
                <TableHead>November 2025</TableHead>
                <TableHead>December 2025</TableHead>
                <TableHead>Januari 2026</TableHead>
                <TableHead>Aangemaakt</TableHead>
              </TableRow>
            </TableHeader>
            
            <TableBody>
             
              <TableRow>
                <TableCell>Test Bedrijf</TableCell>
                <TableCell>
                  <code className="text-xs bg-gray-100 px-2 py-1 rounded">
                    ak_1234567890abcdef
                  </code>
                </TableCell>
                <TableCell>Actief</TableCell>
                <TableCell>50 requests</TableCell>
              </TableRow>
            
            </TableBody>
          </Table>




        </CardContent>
      </Card>

      </div>
    </div>
  )
}